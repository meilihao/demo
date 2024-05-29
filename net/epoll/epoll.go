//go:build linux
// +build linux

// https://github.com/panjf2000/gnet/blob/master/internal/netpoll/epoll_default_poller.go

package main

import (
	"fmt"
	"net"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

func NewEpollM() *EpollM {
	return &EpollM{
		conn: make(map[int]*ServerConn),
	}
}

type EpollM struct {
	look sync.RWMutex
	conn map[int]*ServerConn

	socketFd int //监听socket的fd
	epollFd  int //epoll的fd
}

// 关闭所有的链接
func (e *EpollM) Close() {
	syscall.Close(e.socketFd)
	syscall.Close(e.epollFd)
	for _, con := range e.conn {
		con.Close()
	}
}

// 获取一个链接
func (e *EpollM) GetConn(fd int) *ServerConn {
	e.look.RLock()
	defer e.look.RUnlock()
	return e.conn[fd]
}

// 添加一个链接
func (e *EpollM) AddConn(conn *ServerConn) {
	e.look.Lock()
	e.conn[conn.fd] = conn
	e.look.Unlock()

}

// 删除一个链接
func (e *EpollM) DelConn(fd int) {
	e.look.Lock()
	delete(e.conn, fd)
	e.look.Unlock()
}

// 开启监听
func (e *EpollM) Listen(ipAddr string, port int) error {
	//使用系统调用,打开一个socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}

	//ip地址转换
	var addr [4]byte
	copy(addr[:], net.ParseIP(ipAddr).To4())
	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	})
	if err != nil {
		return err
	}

	//开启监听
	err = syscall.Listen(fd, 10)
	if err != nil {
		return err
	}
	e.socketFd = fd
	return nil
}

// 在死循环中等待client发来的链接
func (e *EpollM) Accept() error {
	for {
		nfd, _, err := syscall.Accept(e.socketFd)
		if err != nil {
			return err
		}
		err = e.EpollAddEvent(nfd)
		if err != nil {
			return err
		}
		e.AddConn(&ServerConn{
			fd: nfd,
		})

	}
}

// 关闭指定的链接
func (e *EpollM) CloseConn(fd int) error {
	conn := e.GetConn(fd)
	if conn == nil {
		return nil
	}
	if err := e.EpollRemoveEvent(fd); err != nil {
		return err
	}
	conn.Close()
	e.DelConn(fd)
	return nil
}

// 创建一个epoll
func (e *EpollM) CreateEpoll() error {
	//通过系统调用,创建一个epoll
	fd, err := syscall.EpollCreate(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return err
	}
	e.epollFd = fd
	return nil
}

func (e *EpollM) CreateEpoll2() error {
	fd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return err
	}
	e.epollFd = fd
	return nil
}

// 处理epoll
func (e *EpollM) HandlerEpoll() error {
	events := make([]syscall.EpollEvent, 100) // make([]unix.EpollEvent, 100)
	//在死循环中处理epoll
	for {
		//msec -1,会一直阻塞,直到有事件可以处理才会返回, n 事件个数
		n, err := syscall.EpollWait(e.epollFd, events, -1) // unix.EpollWait(e.fd, events, msec)
		if err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			//先在map中是否有这个链接
			conn := e.GetConn(int(events[i].Fd))
			if conn == nil { //没有这个链接,忽略
				continue
			}
			if events[i].Events&syscall.EPOLLHUP == syscall.EPOLLHUP || events[i].Events&syscall.EPOLLERR == syscall.EPOLLERR {
				//断开||出错
				if err := e.CloseConn(int(events[i].Fd)); err != nil {
					return err
				}
			} else if events[i].Events == syscall.EPOLLIN {
				//可读事件
				conn.Read()
			}
		}
	}
}

// 向epoll中加事件

/*
//水平触发
evt.events = EPOLLIN;    // LT 水平触发 (默认) EPOLLLT
evt.data.fd = pfd[0];

//边沿触发
evt.events = EPOLLIN | EPOLLET;    // ET 边沿触发
evt.data.fd = pfd[0];

func socketFD(conn net.Conn) int { // get fd from net.Conn
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
*/

func (e *EpollM) EpollAddEvent(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(fd),
		Pad:    0,
	})
}

func (e *EpollM) EpollAddEvent2(fd int) error {
	return unix.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: unix.POLLIN | unix.POLLHUP,
		Fd:     int32(fd),
	})
}

// 从epoll中删除事件
func (e *EpollM) EpollRemoveEvent(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (e *EpollM) EpollRemoveEvent2(fd int) error {
	return unix.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
}

type ServerConn struct {
	fd int
}

// 读取数据
func (s *ServerConn) Read() {
	data := make([]byte, 100)

	//通过系统调用,读取数据,n是读到的长度
	n, err := syscall.Read(s.fd, data)
	if n == 0 {
		return
	}
	if err != nil {
		fmt.Printf("fd %d read error:%s\n", s.fd, err.Error())
	} else {
		fmt.Printf("%d say: %s \n", s.fd, data[:n])
		s.Write([]byte(fmt.Sprintf("hello %d", s.fd)))
	}
}

// 向这个链接中写数据
func (s *ServerConn) Write(data []byte) {
	_, err := syscall.Write(s.fd, data)
	if err != nil {
		fmt.Printf("fd %d write error:%s\n", s.fd, err.Error())
	}
}

// 关闭这个链接
func (s *ServerConn) Close() {
	err := syscall.Close(s.fd)
	if err != nil {
		fmt.Printf("fd%d close error:%s\n", s.fd, err.Error())
	}
}
