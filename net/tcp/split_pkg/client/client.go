// 客户端发送封包
/*
协议:
1. 0x68+len(body)+body
1. min(pkg)=6 && 如果pkg长度超过6, 最后一个byte=len(body)
*/
package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
)

func send(conn net.Conn) {
	var n int32
	for i := 0; i < 1000; i++ {
		n = rand.Int31n(255)
		if n < 6 {
			n = 6
		}

		data := make([]byte, n)
		data[0] = 0x68
		data[1] = byte(n) - 2
		if n > 6 {
			data[len(data)-1] = data[1]
		}
		conn.Write(data)
		//log.Println(i, data)
	}
}

func sender(server string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	defer conn.Close()
	send(conn)
	fmt.Println("connect success")
}

func main() {
	server := "127.0.0.1:9988"

	sender(server)
}
