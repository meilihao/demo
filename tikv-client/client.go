package main

import (
	"context"
	"fmt"

	"github.com/tikv/client-go/config"
	"github.com/tikv/client-go/rawkv"
)

func main() {
	cli, err := rawkv.NewClient(context.TODO(), []string{"127.0.0.1:2379"}, config.Default())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	fmt.Printf("cluster ID: %d\n", cli.ClusterID())

	var key []byte
	keyPrefixs := []string{"aa", "pp", "zz"}
	for _, pre := range keyPrefixs {
		for i := 0; i < 200; i += 10 {
			// put key into tikv
			key = []byte(fmt.Sprintf("%s-%03d", pre, i))
			err = cli.Put(context.TODO(), key, key)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Successfully put %s:%s to tikv\n", key, key)
		}
	}
	fmt.Printf("generate datas done")

	keys, vals, err := cli.Scan(context.TODO(), nil, nil, 1000)
	if err != nil {
		panic(err)
	}
	for idx := range keys {
		fmt.Printf("found kv: %s = %s\n", keys[idx], vals[idx])
	}
	fmt.Printf("get all datas done")

	keys, vals, err = cli.Scan(context.TODO(), []byte("aa-100"), []byte("pp-100"), 1000)
	if err != nil {
		panic(err)
	}
	for idx := range keys {
		fmt.Printf("found kv: %s = %s\n", keys[idx], vals[idx])
	}
	fmt.Printf("get range datas done")

	err = cli.DeleteRange(context.TODO(), []byte("aa-000"), []byte("zz-999"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("clear all datas done")

	// --- raw code
	key = []byte("Company")
	val := []byte("PingCAP")

	// put key into tikv
	err = cli.Put(context.TODO(), key, val)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully put %s:%s to tikv\n", key, val)

	// get key from tikv
	val, err = cli.Get(context.TODO(), key)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found val: %s for key: %s\n", val, key)

	// delete key from tikv
	err = cli.Delete(context.TODO(), key)
	if err != nil {
		panic(err)
	}
	fmt.Printf("key: %s deleted\n", key)

	// get key again from tikv
	val, err = cli.Get(context.TODO(), key)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found val: %s for key: %s\n", val, key)
}
