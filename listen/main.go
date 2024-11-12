package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"strings"
	"tester/vars"
	"time"
)

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o listener listen/main.go && scp listener $apple:/home/ubuntu
func main() {
	// 创建10w个链接到sse上
	// 随机监听15～20分钟

	var start = time.Now()

	var batch = func(red *redis.Client, tks []int64) {
		fmt.Println(len(tks), tks[0])
		_, err := red.Pipelined(vars.BG, func(pipeliner redis.Pipeliner) error {
			for _, i := range tks {
				pipeliner.Set(vars.BG, fmt.Sprintf("pai:user_token:TestSSE%d", i), fmt.Sprintf(`{"id":%d}`, i), 20*time.Minute)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	var uids = make([]int64, 0)
	var i int64
	for i = vars.UIDPrefix; i < vars.Num+vars.UIDPrefix; i++ {
		uids = append(uids, i)
		if len(uids) > 200 {
			batch(vars.Red, uids)
			uids = make([]int64, 0)
		}
	}
	if len(uids) > 0 {
		batch(vars.Red, uids)
	}

	var middle = time.Now()

	fmt.Printf("时长 %ds [%s~%s]\n", middle.Unix()-start.Unix(), start.Format(time.DateTime), middle.Format(time.DateTime))

	var counter = make(chan struct{}, 100000)
	for i = vars.UIDPrefix; i < vars.Num+vars.UIDPrefix; i++ {
		go connect(counter, 300+rand.Int64N(30), i)
		time.Sleep(time.Millisecond)
	}

	for i := 0; i < cap(counter); i++ {
		counter <- struct{}{}
	}
	fmt.Println("完全退出")
}

func connect(counter chan struct{}, second, id int64) {
	counter <- struct{}{}
	defer func() { <-counter }()

	var ids = strconv.FormatInt(id, 10)

	req, err := http.NewRequest("GET", vars.URL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer TestSSE"+ids)

	// 使用 cancel 退出
	//ctx, cancel := context.WithCancel(context.Background())
	//go func() {
	//	time.Sleep(2 * time.Second)
	//	cancel()
	//}()
	//req.WithContext(ctx)

	// 使用timeout退出
	client := &http.Client{Timeout: time.Second * time.Duration(second)}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(id, err)
	}
	defer func(b io.ReadCloser) { _ = b.Close() }(resp.Body)

	reader := bufio.NewReader(resp.Body)

	// 读取SSE消息
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("直接关闭 " + ids)
			break
		}
		if strings.HasPrefix(line, "<html>") {
			fmt.Println("服务器500" + ids)
			break
		}
		if line == "\n" {
			continue
		}
		if err != nil {
			var ne net.Error
			ok := errors.As(err, &ne)
			if ok {
				if ne.Timeout() {
					fmt.Println("时间到了关闭 " + ids)
				} else {
					fmt.Println("未知错误 "+ids, ne.Error())
				}
			} else {
				fmt.Println("收到错误 "+ids, err.Error())
			}
			break
		}
		fmt.Println(ids, strings.TrimRight(line, "\n"))
	}
}
