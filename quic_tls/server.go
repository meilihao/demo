// [tls.RequireAndVerifyClientCert不起作用](https://github.com/lucas-clemente/quic-go/issues/1366)

package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/lucas-clemente/quic-go"
	"go.uber.org/zap"
)

var (
	sugar *zap.SugaredLogger
)

func init() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	sugar = logger.Sugar()
}

func main() {
	pool := x509.NewCertPool()

	caCrt, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		sugar.Panic(err)
		return
	}
	pool.AppendCertsFromPEM(caCrt)

	cert, err := tls.LoadX509KeyPair("server.pem", "server-key.pem")
	if err != nil {
		sugar.Debug(err)
		return
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	ln, err := quic.ListenAddr(":4443", config, nil)
	if err != nil {
		sugar.Debug(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			sugar.Debug(err)
			continue
		}
		go handleConn(conn)
	}
}
func handleConn(conn quic.Session) {
	defer conn.Close()

	stream, err := conn.AcceptStream()
	if err != nil {
		sugar.Debug(err)
		return
	}

	r := bufio.NewReader(stream)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			sugar.Debug(err)
			return
		}
		println(msg)
		n, err := stream.Write([]byte("world\n"))
		if err != nil {
			sugar.Debug(n, err)
			return
		}
	}
}
