package vars

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
)

const (
	URL       = "http://192.168.10.99:8081/listen"
	UIDPrefix = 1000000000
	Num       = 55000
)

var (
	RetryLimitExceededErr = errors.New("retry limit exceeded")
	Red                   = redis.NewClient(&redis.Options{Addr: "192.168.10.99:6379", Password: "123456", DB: 0})
	BG                    = context.Background()
)
