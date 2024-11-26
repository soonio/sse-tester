package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

func main() {
	var nn = suffix()
	fmt.Println(nn)
}

func ip(waiter *sync.WaitGroup, url string, ips *[]byte) {
	defer waiter.Done()
	rsp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[%s] err %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		return
	}
	bs, err := io.ReadAll(rsp.Body)
	if err != nil {
		fmt.Printf("[%s] err %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		return
	}
	*ips = bs
	return
}

func suffix() int64 {
	var ip1, ip2 []byte
	var waiter sync.WaitGroup

	waiter.Add(1)
	go ip(&waiter, "https://ifconfig.me/ip", &ip1)
	waiter.Add(1)
	go ip(&waiter, "https://api64.ipify.org", &ip2)
	waiter.Wait()

	if string(ip1) == string(ip2) && len(ip1) > 0 {
		var ss [32]byte
		for i := 0; i < min(32, len(ip1)); i++ {
			ss[i] = ip1[i]
		}
		return rand.New(rand.NewChaCha8(ss)).Int64N(999)
	}

	return 100 + rand.Int64N(999)
}
