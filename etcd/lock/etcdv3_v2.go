package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type MutexV2 struct {
	key    string
	value  string // The identity of the caller
	ttl    int
	client *clientv3.Client
	s      *concurrency.Session // 会话表示在客户端的生存期内保持活动的租约
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
		m.close()

		return nil, err
	}

	m.mutex = concurrency.NewMutex(m.s, m.key)
	if err = m.mutex.TryLock(context.TODO()); err != nil { // 使用Lock()时, 假设m1已持有锁, m2调用Lock()时会阻塞, 直到m1释放锁
		m.close()

		return nil, err
	}

	log.Printf("locked %s\n", m.value)

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

	log.Printf("unlock %s\n", m.value)

	return nil
}

// NewWithRetry2 trylock with timeout
// n=-1, trylock forever; otherwise trylock n times, if failed, sleep 1s, then try again
func NewWithRetry2(key string, ttl int, endpoints []string, n int64) (*MutexV2, error) {
	var l *MutexV2
	var err error

	i := n
	if i == -1 {
		i = math.MaxInt32
	}
	for i > 0 {
		// log.Printf("Trying to create a node : key=%v\n", key)
		l, err = New2(key, ttl, endpoints)
		log.Printf("%+v\n", err)
		if err == nil {
			return l, nil
		}

		if err != concurrency.ErrLocked {
			return nil, err
		}

		time.Sleep(time.Second)
		i--
	}

	return nil, &LockTimeoutError{Timeout: n}
}
