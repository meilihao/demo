// from https://colobu.com/2017/10/11/badger-a-performant-k-v-store/
package main

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v3"
)

func main() {
	// opt := badger.DefaultOptions("").WithInMemory(true) // use In-Memory Mode/Diskless Mode
	opt := badger.DefaultOptions("./data") // default, persisted to the disk
	db, err := badger.Open(opt)            // It will be created if it doesn't exist.
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// set: read-write transaction
	err = db.Update(func(txn *badger.Txn) error {
		var err error
		if err = txn.Set([]byte("answer"), []byte("42")); err != nil {
			return err
		}

		e := badger.NewEntry([]byte("answer2"), []byte("2"))
		err = txn.SetEntry(e)
		return err
	})
	if err != nil {
		panic(err)
	}

	// get: read-only transaction, no write and no delete
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("answer")) // Txn.Get() returns ErrKeyNotFound if the value is not found
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		fmt.Printf("The answer is: %s\n", val)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// DB.View() 和 DB.Update() 是 DB.NewTransaction() 的 wrapper, 只读事务用Txn.Discard() 即可； 读写事务需要Txn.Commit().
	updates := map[string]string{
		"a": "a",
		"b": "b",
	}
	txn := db.NewTransaction(true) // tree is for read-write transaction
	for k, v := range updates {
		if err := txn.Set([]byte(k), []byte(v)); err == badger.ErrTxnTooBig { // need to limit transaction by commit.
			_ = txn.Commit()
			txn = db.NewTransaction(true)
			_ = txn.Set([]byte(k), []byte(v))
		}
	}
	_ = txn.Commit()

	// iterate
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		it.Close()
		return nil
	})
	if err != nil {
		panic(err)
	}
	// Prefix scans
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		prefix := []byte("ans")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		it.Close()
		return nil
	})
	if err != nil {
		panic(err)
	}
	// iterate keys
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			fmt.Printf("key=%s\n", k)
		}
		it.Close()
		return nil
	})
	// delete
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("answer"))
	})
	if err != nil {
		panic(err)
	}
}
