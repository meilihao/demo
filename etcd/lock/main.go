package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	// testV1()
	testV2()
}

func testV2() {
	// l, err := New2("/mylock", 2, []string{"http://127.0.0.1:2379"})
	// if err != nil {
	// 	log.Printf("Lock failed: ", err)
	// } else {
	// 	log.Printf("Lock OK")
	// }

	// log.Printf("Get the lock. Do something here.")

	// time.Sleep(5 * time.Second)

	// err = l.Unlock()
	// if err != nil {
	// 	log.Printf("Unlock failed", err)
	// } else {
	// 	log.Printf("Unlock OK")
	// }

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		l, err := NewWithRetry2("/mylock", 2, []string{"http://127.0.0.1:2379"}, -1)
		if err != nil {
			fmt.Println("groutine1抢锁失败")
			fmt.Println(err)
			wg.Done()
			return
		}
		defer wg.Done()

		fmt.Println("groutine1抢锁成功")
		time.Sleep(5 * time.Second)
		l.Unlock()
	}()

	//groutine2
	go func() {
		l, err := NewWithRetry2("/mylock", 10, []string{"http://127.0.0.1:2379"}, 2)
		if err != nil {
			fmt.Println("groutine2抢锁失败")
			fmt.Println(err)
			wg.Done()
			return
		}
		defer wg.Done()

		fmt.Println("groutine2抢锁成功")
		l.Unlock()
	}()

	wg.Wait()
	fmt.Println("all done")
}

func testV1() {
	l, err := New("/mylock", 10, []string{"http://127.0.0.1:2379"})
	if err != nil {
		log.Printf("Lock failed: ", err)
	} else {
		log.Printf("Lock OK")
	}

	log.Printf("Get the lock. Do something here.")

	err = l.Unlock()
	if err != nil {
		log.Printf("Unlock failed", err)
	} else {
		log.Printf("Unlock OK")
	}

	// time.Sleep(5 * time.Second)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		l, err := NewWithRetry("/mylock", 2, []string{"http://127.0.0.1:2379"}, -1)
		if err != nil {
			fmt.Println("groutine1抢锁失败")
			fmt.Println(err)
			wg.Done()
			return
		}
		defer wg.Done()

		fmt.Println("groutine1抢锁成功")
		time.Sleep(5 * time.Second)
		l.Unlock()
	}()

	//groutine2
	go func() {
		l, err := NewWithRetry("/mylock", 10, []string{"http://127.0.0.1:2379"}, 2)
		if err != nil {
			fmt.Println("groutine2抢锁失败")
			fmt.Println(err)
			wg.Done()
			return
		}
		defer wg.Done()

		fmt.Println("groutine2抢锁成功")
		l.Unlock()
	}()

	wg.Wait()
	fmt.Println("all done")
}
