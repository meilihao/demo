// from https://yemilice.com/2019/12/13/etcd%E5%88%86%E5%B8%83%E5%BC%8F%E9%94%81%E5%AE%9E%E7%8E%B0%E9%80%89%E4%B8%BB%E6%9C%BA%E5%88%B6-golang/
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultTTL             = 60
	formatErrTryLockFailed = "trylock failed, corrent owner is %s"
	formatTime             = "2006-01-02_15:04:05.999999999"
)

type LockedByOtherError struct {
	Owner string
}

func (e *LockedByOtherError) Error() string {
	return fmt.Sprintf(formatErrTryLockFailed, e.Owner)
}

type LockTimeoutError struct {
	Timeout int64
}

func (e *LockTimeoutError) Error() string {
	return fmt.Sprintf("lock timeout: %ds", e.Timeout)
}

// A Mutex is a mutual exclusion lock which is distributed across a cluster.
type Mutex struct {
	key     string
	value   string // The identity of the caller
	ttl     int64
	client  *clientv3.Client
	lease   clientv3.Lease
	leaseID clientv3.LeaseID
	mutex   *sync.Mutex
	logger  io.Writer
}

// New creates a Mutex with the given key which must be the same
// across the cluster nodes.
// endpoints are the ectd cluster addresses
func New(key string, ttl int64, endpoints []string) (*Mutex, error) {
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

	m := &Mutex{
		key:    key,
		value:  fmt.Sprintf("%v#%v#%v", hostname, os.Getpid(), time.Now().Format(formatTime)), // lock owner
		ttl:    ttl,
		client: cli,
		lease:  clientv3.NewLease(cli), // 上锁用（创建租约，自动续租）
		mutex:  new(sync.Mutex),
		logger: os.Stderr,
	}

	// 设置ttl秒租约（过期时间）
	leaseResp, err := m.lease.Grant(context.TODO(), m.ttl)
	if err != nil {
		m.lease.Close()
		cli.Close()
		return nil, err
	}
	// 租约id
	m.leaseID = leaseResp.ID

	if _, err = m.lease.KeepAlive(context.TODO(), m.leaseID); err != nil {
		m.lease.Close()
		cli.Close()
		return nil, err
	}

	if err = m.lock(); err != nil {
		m.lease.Close()
		cli.Close()
		return nil, err
	}

	return m, nil
}

func (m *Mutex) close() {
	if m.lease != nil {
		m.lease.Close()
	}
	if m.client != nil {
		m.client.Close()
	}
}

// NewWithRetry trylock with timeout
// n=-1, trylock forever; otherwise trylock n times, if failed, sleep 1s, then try again
func NewWithRetry(key string, ttl int64, endpoints []string, n int64) (*Mutex, error) {
	var ok bool
	var l *Mutex
	var err error

	i := n
	if i == -1 {
		i = math.MaxInt32
	}
	for i > 0 {
		// log.Printf("Trying to create a node : key=%v\n", key)
		l, err = New(key, ttl, endpoints)
		if err == nil {
			return l, nil
		}

		if _, ok = err.(*LockedByOtherError); !ok {
			return nil, err
		}

		time.Sleep(time.Second)
		i--
	}

	return nil, &LockTimeoutError{Timeout: n}
}

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *Mutex) lock() (err error) {
	m.mutex.Lock()
	m.debug("Trying to create a node : key=%v", m.key)

	// txn事务：if else then
	txn := clientv3.NewKV(m.client).Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision(m.key), "=", 0)). // 比较key的revision为0(0标示没有key)
										Then(clientv3.OpPut(m.key, m.value, clientv3.WithLease(m.leaseID))).
										Else(clientv3.OpGet(m.key)) // OpGet后TxnResponse才包含Responses

	var resp *clientv3.TxnResponse
	resp, err = txn.Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded { //判断txn.if条件是否成立, 即判断是否抢到了锁
		return &LockedByOtherError{Owner: string(resp.Responses[0].GetResponseRange().Kvs[0].Value)}
	}

	return err
}

// Unlock unlocks m.
// It is a runtime error if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *Mutex) Unlock() (err error) {
	defer m.mutex.Unlock()
	defer m.close()

	var resp *clientv3.LeaseRevokeResponse
	resp, err = m.lease.Revoke(context.TODO(), m.leaseID) // 终止租约（去掉过期时间）, key will deleted auto. ps: 租约提前过期时该调用也不报错
	if err == nil {
		m.debug("Delete %v OK", m.key)
		return nil
	} else {
		m.debug("Delete %v failed: %v", m.key, err)
	}

	_ = resp

	return err
}

func (m *Mutex) debug(format string, v ...interface{}) {
	if m.logger != nil {
		m.logger.Write([]byte(m.value))
		m.logger.Write([]byte(" "))
		m.logger.Write([]byte(fmt.Sprintf(format, v...)))
		m.logger.Write([]byte("\n"))
	}
}

func (m *Mutex) SetDebugLogger(w io.Writer) {
	m.logger = w
}
