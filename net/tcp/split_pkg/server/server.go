// 服务端解包过程
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	_ "net/http/pprof"
)

func main() {
	netListen, err := net.Listen("tcp", ":9988")
	CheckError(err)

	defer netListen.Close()

	//go http.ListenAndServe("0.0.0.0:6060", nil)

	log.Println("Waiting for clients")
	for {
		conn, err := netListen.Accept()
		if err != nil {
			continue
		}

		log.Println(conn.RemoteAddr().String(), " tcp connect success")
		go handleConnectionWithReader(conn)
		//go handleConnectionWithScanner(conn)
	}
}

// 推荐:
func handleConnectionWithReader(conn net.Conn) {
	var err error
	var peekBuf []byte
	var pkgLen byte
	reader := bufio.NewReaderSize(conn, 4096)
	var n int64
	for {
		if peekBuf, err = reader.Peek(2); err != nil {
			if err == io.EOF {
				log.Println("peek header remote connect closed, %v", err)
			} else {
				log.Fatalf("peek header receive failed, %v", err)
			}

			return
		}

		if peekBuf[0] != 0x68 {
			log.Fatalf("receive apdu peek, %v", peekBuf)
			return
		}

		pkgLen = 2 + peekBuf[1]
		if peekBuf, err = reader.Peek(int(pkgLen)); err != nil {
			if err == io.EOF {
				log.Println("peek pkg remote connect closed, %v", err)
			} else {
				log.Fatalf("peek pkg receive failed, %v", err)
			}

			return
		}

		rawData := make([]byte, pkgLen)
		if _, err = io.ReadFull(reader, rawData); err != nil {
			if err == io.EOF {
				log.Println("remote connect closed, %v", err)
			} else {
				log.Fatalf("receive failed, %v", err)
			}

			return
		}

		n += 1
		log.Println("RX Raw %d", n)
		if pkgLen > 6 {
			if rawData[len(rawData)-1] != rawData[1] {
				panic(rawData)
			}
		}
	}
}

func packetSlitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF && len(data) >= 6 {
		if data[0] != 0x68 {
			return 0, nil, fmt.Errorf("receive apdu peek, %v", data[:6])
		}

		if len(data) < 2+int(data[1]) {
			return 0, nil, nil
		}

		pl := 2 + int(data[1])
		return pl, data[:pl], nil
	}
	return
}

// https://juejin.cn/post/6844903882108174343
// **不推荐**: 如果客户端只是正常调用 conn.Close()，没有发送 EOF 或触发读取错误, scanner.Scan() 就会一直阻塞. 临时解决方法: server设置 ReadDeadline 或client用CloseWrite() 发 EOF
func handleConnectionWithScanner(conn net.Conn) {
	// 重置 deadline 每次读取
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	var err error
	scanner := bufio.NewScanner(conn)
	scanner.Split(packetSlitFunc)
	var n int64
	for scanner.Scan() {
		// 重置 deadline 每次读取
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		buf := scanner.Bytes()
		rawData := make([]byte, len(buf))
		copy(rawData, buf)

		n += 1
		log.Println("RX Raw %d", n)
		if len(rawData) > 6 {
			if rawData[len(rawData)-1] != rawData[1] {
				panic(rawData)
			}
		}
	}
	if err = scanner.Err(); err != nil {
		log.Println("remote connect closed, %v", err)
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
