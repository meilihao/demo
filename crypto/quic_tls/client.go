package main

import (
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
	cert, err := tls.LoadX509KeyPair("client.pem", "client-key.pem")
	if err != nil {
		sugar.Panic(err)
		return
	}

	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		sugar.Panic(err)
		return
	}
	pool.AppendCertsFromPEM(caCrt)

	_ = cert
	conf := &tls.Config{
		RootCAs: pool,
		//Certificates: []tls.Certificate{cert},
	}

	session, err := quic.DialAddr("127.0.0.1:4443", conf, nil)
	if err != nil {
		sugar.Debug(err)
		return
	}
	defer session.Close()

	stream, err := session.OpenStreamSync()
	if err != nil {
		sugar.Debug(err)
		return
	}

	n, err := stream.Write([]byte("hello\n"))
	if err != nil {
		sugar.Debug(n, err)
		return
	}

	buf := make([]byte, 100)
	n, err = stream.Read(buf)
	if err != nil {
		sugar.Debug(n, err)
		return
	}
	println(string(buf[:n]))
}
