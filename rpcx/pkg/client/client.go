package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"example/pkg/pb"

	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"
)

var (
	etcdAddr = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath = flag.String("base", "/rpcx_test", "prefix path")
)

func main() {
	flag.Parse()

	d, _ := etcd_client.NewEtcdV3Discovery(*basePath, "Arith", []string{*etcdAddr}, nil)

	//sendRPC(d)
	sendFile(d)
}

func sendRPC(d client.ServiceDiscovery) {
	xclient := client.NewXClient("Arith", client.Failover, client.RoundRobin, d, client.DefaultOption)
	defer xclient.Close()

	args := &pb.ProtoArgs{
		A: 10,
		B: 20,
	}

	for {
		reply := &pb.ProtoReply{}
		err := xclient.Call(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Printf("failed to call: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("%d * %d = %d", args.A, args.B, reply.C)

		time.Sleep(5 * time.Second)
	}
}

func sendFile(d client.ServiceDiscovery) {
	xclient := client.NewXClient(share.SendFileServiceName, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xclient.Close()

	f, err := os.Open("../../go.mod")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = xclient.SendFile(context.Background(), "../../go.mod", 0)
	if err != nil {
		panic(err)
	}
	log.Println("send ok")
}
