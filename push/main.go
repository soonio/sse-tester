package main

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"strconv"
	"tester/vars"
	"time"
)

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o pusher push/main.go && scp pusher $apple:/home/ubuntu
func main() {
	var pusher = NewPushLogic(vars.Red, "pai:sse:connect:alive")
	var concurrency = make(chan struct{}, 20)

	var start = time.Now()

	var i int64
	for i = vars.UIDPrefix; i < vars.Num+vars.UIDPrefix; i++ {
		concurrency <- struct{}{}
		go func(id int64) {
			defer func() { <-concurrency }()
			msg, _ := jsoniter.MarshalToString(map[string]any{
				"type":      "err",
				"body":      "这是一句话啊",
				"timestamp": time.Now().Format(time.DateTime),
				"key":       id,
			})

			if ok, err := pusher.Dispatch(vars.BG, strconv.FormatInt(id, 10), msg); err != nil {
				fmt.Println("推送异常", id, err.Error())
			} else {
				if ok {
					fmt.Println("推送成功", id)
				} else {
					fmt.Println("推送失败", id)
				}
			}
			if err := pusher.DispatchWithRetry(vars.BG, strconv.FormatInt(id, 10), msg, 5); err != nil {
				fmt.Println("推送异常", id, err.Error())
			} else {
				fmt.Println("推送成功", id)
			}
		}(i)
	}
	for i := 0; i < cap(concurrency); i++ {
		concurrency <- struct{}{}
	}

	var end = time.Now()

	fmt.Printf("时长 %ds [%s~%s]\n", end.Unix()-start.Unix(), start.Format(time.DateTime), end.Format(time.DateTime))
}

type PushLogic struct {
	red       *redis.Client
	script    *redis.Script
	durations []time.Duration
	channel   string
}

func NewPushLogic(red *redis.Client, channel string) *PushLogic {
	return &PushLogic{
		red:    red,
		script: redis.NewScript(`local ccc = redis.call('GET', KEYS[1]); if ccc then redis.call('PUBLISH', ccc, ARGV[1]);return 1; else return 0 end`),
		durations: []time.Duration{
			500 * time.Millisecond,
			1000 * time.Millisecond,
			1000 * time.Millisecond,
			2000 * time.Millisecond,
			3000 * time.Millisecond,
		},
		channel: channel,
	}
}

// Dispatch 不带重试的推送
func (d *PushLogic) Dispatch(ctx context.Context, uid, msg string) (bool, error) {
	code, err := d.script.Run(ctx, d.red, []string{d.channel + ":" + uid}, uid+"|msg|"+msg).Result()
	if err != nil {
		return false, err
	}
	return code.(int64) == 1, nil
}

// DispatchWithRetry 带有重试的推送
func (d *PushLogic) DispatchWithRetry(ctx context.Context, uid, msg string, attempts int) error {
	attempts = min(attempts, len(d.durations))
	attempts = max(attempts, 0)

	for i := -1; i < attempts; i++ {
		if i != -1 { // 正常的第一跳过
			time.Sleep(d.durations[i])
		}
		ok, err := d.Dispatch(ctx, uid, msg)
		if err != nil { // 有错误，很有可能redis挂了或者脚本有错误，不进行重试
			return err
		}
		if ok {
			return nil
		}
	}
	return vars.RetryLimitExceededErr
}

func (d *PushLogic) Push(uid int64, msg string) error {
	return d.DispatchWithRetry(context.Background(), strconv.FormatInt(uid, 10), msg, 7)
}
