package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
)

var (
	addr             = flag.String("addr", "localhost:8972", "server address")
	etcdAddr         = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath         = flag.String("base", "/rpcx_test", "prefix path")
	fileTransferAddr = flag.String("transfer-addr", "localhost:8973", "data transfer address")
)

func main() {
	flag.Parse()

	s := server.NewServer()
	addRegistryPlugin(s)

	p := server.NewFileTransfer(*fileTransferAddr, saveFilehandler, nil, 1000)
	s.EnableFileTransfer(share.SendFileServiceName, p)

	s.RegisterName("Arith", new(Arith), "")
	err := s.Serve("tcp", *addr)
	if err != nil {
		panic(err)
	}
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}

func saveFilehandler(conn net.Conn, args *share.FileTransferArgs) {
	fmt.Printf("received file name: %s, size: %d\n", args.FileName, args.FileSize)
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Printf("error read: %v\n", err)
		return
	}
	fmt.Printf("file content: %s\n", data)
}
