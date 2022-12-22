// [MySQL · RocksDB · TransactionDB 介绍](https://developer.aliyun.com/article/655903)
package main

import (
	"os"
	"strings"
	"testing"

	"github.com/linxGnu/grocksdb"
	"github.com/stretchr/testify/assert"
)

var (
	dbPathOT = "/tmp/rocksdb_optimistic_transaction_example"
	dbPathT  = "/tmp/rocksdb_transaction_example"
)

// https://github.com/facebook/rocksdb/blob/main/java/samples/src/main/java/OptimisticTransactionSample.java
func TestOptimisticTransactionSample(t *testing.T) {
	options := grocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)

	// RocksDB的Transaction分为两类：Pessimistic和Optimistic，类似悲观锁和乐观锁的区别，PessimisticTransaction的冲突检测和加锁是在事务中每次写操作之前做的（commit后释放），如果失败则该操作失败；OptimisticTransaction不加锁，冲突检测是在commit阶段做的，commit时发现冲突则失败。

	// [当使用TransactionDB或者OptimisticTransactionDB的时候，RocksDB将支持事务。事务带有简单的BEGIN/COMMIT/ROLLBACK API，并且允许应用并发地修改数据，具体的冲突检查，由Rocksdb来处理。RocksDB支持悲观和乐观的并发控制](https://wanghenshui.github.io/rocksdb-doc-cn/doc/Transactions.html)
	// 一个TransactionDB在由大量并发工作压力的时候，相比OptimisticTransactionDB有更好的表现。然而，由于非常过激的上锁策略，使用TransactionDB会有一定的性能损耗。TransactionDB会在所有写操作的时候做冲突检查，包括不使用事务写入的时候
	// OptimisticTransactionDB在大量非事务写入，而少量事务写入的场景，会比TransactionDB性能更好
	txnDb, err := grocksdb.OpenOptimisticTransactionDb(options, dbPathOT)
	assert.Nil(t, err)

	writeOptions := grocksdb.NewDefaultWriteOptions()
	readOptions := grocksdb.NewDefaultReadOptions()

	readCommittedOT(t, txnDb, writeOptions, readOptions)
	//repeatableRead(txnDb, writeOptions, readOptions)
	//readCommittedMonotonicAtomicViews(txnDb, writeOptions, readOptions)
}

func readCommittedOT(t *testing.T, txnDb *grocksdb.OptimisticTransactionDB, writeOptions *grocksdb.WriteOptions, readOptions *grocksdb.ReadOptions) {
	key1 := []byte("abc")
	value1 := []byte("def")

	//key2 := []byte("xyz")
	//value2 := []byte("zzz")

	to := grocksdb.NewDefaultOptimisticTransactionOptions()
	txn := txnDb.TransactionBegin(writeOptions, to, nil)
	defer txn.Destroy()

	value, _ := txn.Get(readOptions, key1)
	assert.Equal(t, false, value.Exists())

	txn.Put(key1, value1)

	// value, _ = txnDb.Get(readOptions, key1) // ??? not found txnDb.Get
	// assert.Equal(t, false, value.Exists())

	txn.Commit()
}

// go test -timeout 30s -run ^TestTransactionSample$
// https://github.com/facebook/rocksdb/blob/main/java/samples/src/main/java/TransactionSample.java
func TestTransactionSample(t *testing.T) {
	os.RemoveAll(dbPathT)

	options := grocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)

	txnDbOptions := grocksdb.NewDefaultTransactionDBOptions()
	txnDb, err := grocksdb.OpenTransactionDb(options, txnDbOptions, dbPathT)
	assert.Nil(t, err)

	writeOptions := grocksdb.NewDefaultWriteOptions()
	readOptions := grocksdb.NewDefaultReadOptions()

	readCommittedT(t, txnDb, writeOptions, readOptions)
	repeatableReadT(t, txnDb, writeOptions, readOptions)
	readCommittedMonotonicAtomicViews(t, txnDb, writeOptions, readOptions)
}

func readCommittedT(t *testing.T, txnDb *grocksdb.TransactionDB, writeOptions *grocksdb.WriteOptions, readOptions *grocksdb.ReadOptions) {
	key1 := []byte("abc")
	value1 := []byte("def")

	key2 := []byte("xyz")
	value2 := []byte("zzz")

	to := grocksdb.NewDefaultTransactionOptions()
	txn := txnDb.TransactionBegin(writeOptions, to, nil)
	defer txn.Destroy()

	value, _ := txn.Get(readOptions, key1) // when no found, err is nil
	assert.Equal(t, false, value.Exists())

	txn.Put(key1, value1)

	{
		// OUTSIDE this transaction
		// Does not affect txn since this is an unrelated key.
		// If we wrote key 'abc' here, the transaction would fail to commit.
		value, _ = txnDb.Get(readOptions, key1)
		assert.Equal(t, false, value.Exists())

		txnDb.Put(writeOptions, key2, value2)
	}

	txn.Commit()
}

func repeatableReadT(t *testing.T, txnDb *grocksdb.TransactionDB, writeOptions *grocksdb.WriteOptions, readOptions *grocksdb.ReadOptions) {
	key1 := []byte("ghi")
	value1 := []byte("jkl")

	to := grocksdb.NewDefaultTransactionOptions()
	to.SetSetSnapshot(true)

	txn := txnDb.TransactionBegin(writeOptions, to, nil)
	defer txn.Destroy()

	snapshot := txn.GetSnapshot()

	{
		// UTSIDE of transaction
		txnDb.Put(writeOptions, key1, value1)
	}

	// Attempt to read a key using the snapshot.  This will fail since
	// the previous write outside this txn conflicts with this read.
	readOptions.SetSnapshot(snapshot)

	_, err := txn.GetForUpdate(readOptions, key1)
	assert.Contains(t, strings.ToLower(err.Error()), "busy") // err: Resource busy

	txn.Rollback()

	// Clear snapshot from read options since it is no longer valid
	snapshot.Destroy()
	readOptions.SetSnapshot(snapshot)
}

func readCommittedMonotonicAtomicViews(t *testing.T, txnDb *grocksdb.TransactionDB, writeOptions *grocksdb.WriteOptions, readOptions *grocksdb.ReadOptions) {
	// keyX := []byte("x")
	// valueX := []byte("x")

	// keyY := []byte("y")
	// valueY := []byte("y")

	// to := grocksdb.NewDefaultTransactionOptions()
	// to.SetSetSnapshot(true)

	// txn := txnDb.TransactionBegin(writeOptions, to, nil)
	// defer txn.Destroy()

	// snapshot := txnDb.GetSnapshot() // no txnDb.GetSnapshot

	// ...
}
