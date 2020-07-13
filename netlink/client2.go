package main

import (
	"os"

	"github.com/mdlayher/netlink"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func init() {
	log.SetReportCaller(true)
}

var NETLINK_TEST int = 30
var NETLINK_PORT uint32 = 100

func main() {
	fd, err := unix.Socket(
		// Always used when opening netlink sockets.
		// 打开 netlink 套接字时始终使用。
		unix.AF_NETLINK,
		// Seemingly used interchangeably with SOCK_DGRAM,
		// but it appears not to matter which is used.
		// 似乎与 SOCK_DGRAM 可互换使用，使用哪个并不重要。
		unix.SOCK_RAW,
		// The netlink family that the socket will communicate
		// with, such as NETLINK_ROUTE or NETLINK_GENERIC.
		// 套接字与之通信的 netlink 系列，如 NETLINK_ROUTE 或 NETLINK_GENERIC。
		NETLINK_TEST, // 会与unix.Recvfrom中收到消息的消息头中的Type对应
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		unix.Close(fd)
	}()

	err = unix.Bind(fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: 0,
		Pid:    NETLINK_PORT,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Info("create fd sucess")

	log.Info("start send msg")
	err = unix.Sendto(fd, buildMsg("hello meilihao!"), 0, &unix.SockaddrNetlink{
		// Always used when sending on netlink sockets.
		Family: unix.AF_NETLINK,
		Pid:    0, // to kernel
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Info("send msg done")

	log.Info("start receive msg")
	b := make([]byte, os.Getpagesize())
	for {
		// Peek at the buffer to see how many bytes are available.
		n, _, err := unix.Recvfrom(fd, b, unix.MSG_PEEK)
		if err != nil {
			log.Error(err)

			continue
		}
		// Break when we can read all messages.
		if n < len(b) {
			break
		}
		// Double in size if not enough bytes.
		b = make([]byte, len(b)*2)
	}
	// Read out all available messages.
	n, _, _ := unix.Recvfrom(fd, b, 0)
	log.Infof("get msg len: %d\n", n)

	log.Infof("data: %+v", b[:n])

	m := &netlink.Message{}
	if err = m.UnmarshalBinary(b[:n]); err != nil { // mdlayher/netlink要求int(m.Header.Length) == n, 与实际情况不符所以报错. 因此自行按照m.UnmarshalBinary源码重新实现
		log.Fatal(err)
	}

	log.Info("receive msg done")
}

func buildMsg(content string) []byte {
	b := []byte(content)

	data := make([]byte, nlmsgAlign(len(b)+1))
	copy(data, b)

	msg := netlink.Message{
		Header: netlink.Header{
			// Length of header(16 B), plus payload.
			Length: 16 + uint32(len(b)+1), // 是指有效长度, 不算为对齐而填充的字节
			// Set to zero on requests.
			Type: 0,
			// Indicate that message is a request to the kernel.
			Flags: netlink.Request,
			// Sequence number selected at random.
			Sequence: 1,
			// PID set to process's ID.
			PID: NETLINK_PORT,
		},
		// An arbitrary byte payload. May be in a variety of formats.
		Data: data,
	}

	buf, err := msg.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	return buf
}

const nlmsgAlignTo = 4

// from https://github.com/mdlayher/netlink/blob/master/align.go
func nlmsgAlign(n int) int {
	return (n + nlmsgAlignTo - 1) & ^(nlmsgAlignTo - 1) // (nlmsgAlignTo-1)取反再与(n + nlmsgAlignTo -1)即可减去多余的个数
}
