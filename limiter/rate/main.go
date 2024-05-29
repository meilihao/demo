// [Golang官方限流器的用法详解](https://cloud.tencent.com/developer/article/1847918)
// [Golang 官方限流器time/rate使用](https://piaohua.github.io/post/golang/20200815-golang-rate-limiter/)
package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	speed := 20 * 1 << 20
	maxSpeed := 50 * 1 << 20
	N := 20 * 1 << 20

	// 第一个参数是 r Limit。代表每秒可以向 Token 桶中产生多少 token
	// 第二个参数是 b int。b 代表 Token 桶的容量大小
	limiter := rate.NewLimiter(rate.Limit(speed), maxSpeed)

	n := 20

	fmt.Println("---start:", time.Now())
	for {
		n--
		if n < 1 {
			break
		}

		Reserve(limiter, n, N)
		//Wait(limiter, n, N)
	}
	fmt.Println("---end:", time.Now())
}

// 支持context 的 Deadline 或者 Timeout
func Wait(limiter *rate.Limiter, n, N int) {
	fmt.Println("r", limiter.Tokens())
	err := limiter.WaitN(context.Background(), N)
	fmt.Println("-", limiter.Tokens(), err)
	fmt.Printf("act %02d, %d %v\n", n, limiter.Burst(), time.Now())
	//time.Sleep(time.Second)
}

func Reserve(limiter *rate.Limiter, n, N int) {
	fmt.Println("r", limiter.Tokens())
	r := limiter.ReserveN(time.Now(), N) // 不能使用r.OK(), 因为它仅表示是否拿到了token, 不能所拿token是否足够
	if r.OK() {
		if r.Delay() != 0 {
			fmt.Printf("wait %02d %v\n", n, time.Now())
			time.Sleep(r.Delay())
		}
	} else { // 避免没有取到token时返回rate.InfDuration
		fmt.Printf("wait %02d %v with no token\n", n, time.Now())
		time.Sleep(time.Second)
	}

	fmt.Println("-", limiter.Tokens())
	fmt.Printf("act %02d %v\n", n, time.Now())
}
