package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/clientv3"
)

type MutexV2 struct {
	key    string
	value  string // The identity of the caller
	ttl    int
	client *clientv3.Client
	s      *concurrency.Session
	mutex  *concurrency.Mutex
	logger io.Writer
}

func New2(key string, ttl int, endpoints []string) (*MutexV2, error) {
	if len(key) == 0 || len(endpoints) == 0 {
		return nil, errors.New("wrong lock args")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()

	if key[0] != '/' {
		key = "/" + key
	}

	if ttl < 1 {
		ttl = defaultTTL
	}

	m := &MutexV2{
		key:    key,
		value:  fmt.Sprintf("%v#%v#%v", hostname, os.Getpid(), time.Now().Format(formatTime)), // lock owner
		ttl:    ttl,
		client: cli,
		logger: os.Stderr,
	}

	m.s, err = concurrency.NewSession(cli, concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}

	m.mutex = concurrency.NewMutex(m.s, m.key)
	if err = m.mutex.Lock(context.TODO()); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MutexV2) close() {
	if m.s != nil {
		m.s.Close()
	}
	if m.client != nil {
		m.client.Close()
	}
}

func (m *MutexV2) Unlock() (err error) {
	defer m.close()

	m.mutex.Unlock(context.TODO())
	if _, err := m.client.Delete(context.TODO(), m.key); err != nil {
		return err
	}

	return nil
}
