package main

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func ConnServer(i int) error {
	conn, err := net.Dial("tcp", "127.0.0.1:8088")
	if err != nil {
		return err
	}

	conn.Write([]byte(fmt.Sprintf("hello_%d", i)))

	data := make([]byte, 100)
	n, err := conn.Read(data)
	if err != nil {
		return err
	}
	fmt.Println(string(data[:n]))

	time.Sleep(100 * time.Millisecond)
	return conn.Close()
}

func TestClient(t *testing.T) {
	err := ConnServer(0)
	if err != nil {
		panic(err)
	}
}

func TestClients(t *testing.T) {
	for i := 0; i < 100; i++ {
		err := ConnServer(i)
		if err != nil {
			panic(err)
		}
	}
}

func TestClientMulti(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			ConnServer(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
