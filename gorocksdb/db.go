// rocksdb 6.11.4
package main

import (
	"errors"
	"log"
	"fmt"
	"bytes"

	gorocksdb "github.com/tecbot/gorocksdb"
	//"strconv"
)

const (
	DB_PATH = "gorocksdb.db"
)

func main() {
	log.SetFlags(log.LstdFlags|log.Llongfile)

	db, err := OpenDB()
	if err != nil {
		log.Println("fail to open db,", nil, db)
	}

	readOptions := gorocksdb.NewDefaultReadOptions()
	readOptions.SetFillCache(true)
	defer readOptions.Destroy()

	writeOptions := gorocksdb.NewDefaultWriteOptions()
	writeOptions.SetSync(true)
	defer writeOptions.Destroy()

	err = db.Put(writeOptions, []byte("cf-0"), nil)
	log.Println(err)

	err = db.Put(writeOptions, []byte("cf-1"), []byte(""))
	log.Println(err)


	log.Println("--- key exist")
	raw, err := db.Get(readOptions, []byte("cf-0")) // raw.Exists() is true when value =nil
	log.Println(raw, raw.Exists(), err)

	raw, err = db.Get(readOptions, []byte("cf-1"))
	log.Println(raw, raw.Exists(), err)

	raw, err = db.Get(readOptions, []byte("cf-2"))
	log.Println(raw, raw.Exists(), err)

	log.Println("--- Get/Put")
	for i := 0; i < 20; i += 1 {
		var keyStr string
		
		if i&1==0{
			keyStr = "t2-"
		}else{
			keyStr = "t1-"
		}
		keyStr+=fmt.Sprintf("%03d",i)

		var key []byte = []byte(keyStr)
		db.Put(writeOptions, key, key)
		log.Println(i, keyStr)

		slice, err2 := db.Get(readOptions, key)
		if err2 != nil {
			log.Println("get data exception：", key, err2)
			continue
		}
		log.Println("get data：", slice.Size(), string(slice.Data()))
		slice.Free()
	}

	log.Println("--- range by key prefix")
	readOptions.SetPrefixSameAsStart(true) // 搜索相同前缀
	it := db.NewIterator(readOptions)
	defer it.Close()

	it.Seek([]byte("t1-"))

	endKey:=[]byte("t1-015")
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if bytes.Compare(key.Data(), endKey)>0{
		        log.Println("end key")

		        key.Free()
				value.Free()
		        break
		}

		log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

		key.Free()
		value.Free()
	}
	if err := it.Err(); err != nil {
		log.Println(err)
	}

	log.Println("--- SeekToFirst")
	readOptions.SetPrefixSameAsStart(false) // 设为false, 否则NewIterator只输出与获取到的第一个key相同key prefix的kv
	it= db.NewIterator(readOptions)
	defer it.Close()

	for it.SeekToFirst(); it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

		key.Free()
		value.Free()
	}

	log.Println("--- get start key")
	it = db.NewIterator(readOptions)
	defer it.Close()

	it.Seek([]byte("t1-003"))

	for ; it.Valid(); it.Next() {
		// found
		key := it.Key()
		value := it.Value()

		log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

		key.Free()
		value.Free()
		break
	}
	if err := it.Err(); err != nil {
		log.Println(err)
	}

	log.Println("--- get end key")
	it = db.NewIterator(readOptions)
	defer it.Close()

	it.SeekForPrev([]byte("t1-030"))
	if !it.Valid(){ // not found
		it.SeekToLast()
	}

	for ; it.Valid(); it.Prev() {
		// found
		key := it.Key()
		value := it.Value()

		log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

		key.Free()
		value.Free()
		break
	}
	if err := it.Err(); err != nil {
		log.Println(err)
	}


	log.Println("--- range delete")
	it= db.NewIterator(readOptions)
	defer it.Close()

	for it.SeekToFirst(); it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

		key.Free()
		value.Free()
	}

	wb := gorocksdb.NewWriteBatch()
	defer wb.Destroy()

	dStartKey:=[]byte("t2-005")
	dEndKey:=[]byte("t2-010")
	wb.DeleteRange(dStartKey, dEndKey)

	db.Write(writeOptions, wb)

	it= db.NewIterator(readOptions)
	defer it.Close()

	for it.SeekToFirst(); it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

		key.Free()
		value.Free()
	}
}

// opendb
func OpenDB() (*gorocksdb.DB, error) {
	options := gorocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)

	//options.SetComparator(&MyComparator{}) // 自定义key compare会报错

	bloomFilter := gorocksdb.NewBloomFilter(10)

	readOptions := gorocksdb.NewDefaultReadOptions()
	readOptions.SetFillCache(false)

	rateLimiter := gorocksdb.NewRateLimiter(10000000, 10000, 10)
	options.SetRateLimiter(rateLimiter)
	options.SetCreateIfMissing(true)
	options.EnableStatistics()
	options.SetWriteBufferSize(8 * 1024)
	options.SetMaxWriteBufferNumber(3)
	options.SetMaxBackgroundCompactions(10)
	options.SetCompression(gorocksdb.SnappyCompression)
	options.SetCompactionStyle(gorocksdb.UniversalCompactionStyle)

	options.SetHashSkipListRep(2000000, 4, 4)

	blockBasedTableOptions := gorocksdb.NewDefaultBlockBasedTableOptions()
	blockBasedTableOptions.SetBlockCache(gorocksdb.NewLRUCache(64 * 1024))
	blockBasedTableOptions.SetFilterPolicy(bloomFilter)
	blockBasedTableOptions.SetBlockSizeDeviation(5)
	blockBasedTableOptions.SetBlockRestartInterval(10)
	blockBasedTableOptions.SetBlockCacheCompressed(gorocksdb.NewLRUCache(64 * 1024))
	blockBasedTableOptions.SetCacheIndexAndFilterBlocks(true)
	blockBasedTableOptions.SetIndexType(gorocksdb.KHashSearchIndexType)

	options.SetBlockBasedTableFactory(blockBasedTableOptions)
	//log.Println(bloomFilter, readOptions)
	options.SetPrefixExtractor(gorocksdb.NewFixedPrefixTransform(3)) // 要通过前缀检索，需要首先创建一个前缀提取器. 注意前缀的长度, 要与Put时一致

	options.SetAllowConcurrentMemtableWrites(false)

	db, err := gorocksdb.OpenDb(options, DB_PATH)

	if err != nil {
		log.Fatalln("OPEN DB error", db, err)
		db.Close()
		return nil, errors.New("fail to open db")
	} else {
		log.Println("OPEN DB success", db)
	}
	
	go func(){
		for {
			s:=options.GetStatisticsString()
			fmt.Println(s)
		
			time.Sleep(60 * time.Second)
		}
	}()

	return db, nil
}

type MyComparator struct{}

func (c *MyComparator) Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}

func (c *MyComparator) Name() string {
	return "MyComparator"
}
