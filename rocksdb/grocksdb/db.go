// rocksdb 6.29.5
package main

/*
#cgo CFLAGS="-I/usr/local/include/rocksdb"
#cgo LDFLAGS="-L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd"
*/
import (
	"errors"
	"fmt"
	"log"

	"github.com/linxGnu/grocksdb"
	//"bytes"
	//"time"
	//"strconv"
)

const (
	DB_PATH = "test.db"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	fmt.Println("--start")
	db, err := OpenDB()
	if err != nil {
		log.Println("fail to open db,", nil, db)
	}

	readOptions := grocksdb.NewDefaultReadOptions()
	readOptions.SetFillCache(true)
	defer readOptions.Destroy()

	writeOptions := grocksdb.NewDefaultWriteOptions()
	writeOptions.SetSync(true)
	defer writeOptions.Destroy()

	keyPrefixs := []string{"aa", "pp", "zz"}
	for _, pre := range keyPrefixs {
		for i := 50; i < 20000; i += 1 {
			// put key into tikv
			key := []byte(fmt.Sprintf("%s-%03d", pre, i))
			err = db.Put(writeOptions, key, key)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Successfully put %s:%s\n", key, key)
		}
	}
	fmt.Println("generate datas done")

	// fmt.Println("get all data:")
	// it := db.NewIterator(readOptions)
	// defer it.Close()
	// for it.SeekToFirst(); it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()

	// 	fmt.Printf("%s = %s\n", key.Data(), value.Data())

	// 	key.Free()
	// 	value.Free()
	// }

	var StartKey = func(start []byte) []byte {
		it := db.NewIterator(readOptions)
		defer it.Close()

		var target []byte = make([]byte, len(start))

		it.Seek(start)
		for ; it.Valid(); it.Next() {
			key := it.Key()
			copy(target, key.Data())

			key.Free()
			break
		}

		if it.Err() != nil {
			log.Fatal(it.Err())
		}
		return target
	}

	fmt.Printf("start key: %s\n", StartKey([]byte("pp-011")))

	// err = db.Put(writeOptions, []byte("cf-0"), nil)
	// log.Println(err)

	// err = db.Put(writeOptions, []byte("cf-1"), []byte(""))
	// log.Println(err)

	// log.Println("--- key exist")
	// raw, err := db.Get(readOptions, []byte("cf-0")) // raw.Exists() is true when value =nil
	// log.Println(raw, raw.Exists(), err)

	// raw, err = db.Get(readOptions, []byte("cf-1"))
	// log.Println(raw, raw.Exists(), err)

	// raw, err = db.Get(readOptions, []byte("cf-2"))
	// log.Println(raw, raw.Exists(), err)

	// log.Println("--- Get/Put")
	// for i := 0; i < 20; i += 1 {
	// 	var keyStr string

	// 	if i&1==0{
	// 		keyStr = "t2-"
	// 	}else{
	// 		keyStr = "t1-"
	// 	}
	// 	keyStr+=fmt.Sprintf("%03d",i)

	// 	var key []byte = []byte(keyStr)
	// 	db.Put(writeOptions, key, key)
	// 	log.Println(i, keyStr)

	// 	slice, err2 := db.Get(readOptions, key)
	// 	if err2 != nil {
	// 		log.Println("get data exception：", key, err2)
	// 		continue
	// 	}
	// 	log.Println("get data：", slice.Size(), string(slice.Data()))
	// 	slice.Free()
	// }

	// log.Println("--- range by key prefix")
	// readOptions.SetPrefixSameAsStart(true) // 搜索相同前缀
	// it := db.NewIterator(readOptions)
	// defer it.Close()

	// it.Seek([]byte("t1-"))

	// endKey:=[]byte("t1-015")
	// for ; it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()

	// 	if bytes.Compare(key.Data(), endKey)>0{
	// 	        log.Println("end key")

	// 	        key.Free()
	// 			value.Free()
	// 	        break
	// 	}

	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// }
	// if err := it.Err(); err != nil {
	// 	log.Println(err)
	// }

	// log.Println("--- range by key prefix no end")
	// readOptions.SetPrefixSameAsStart(true) // 搜索相同前缀
	// it = db.NewIterator(readOptions)
	// defer it.Close()

	// it.Seek([]byte("t1-"))

	// for ; it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()

	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// }
	// if err := it.Err(); err != nil {
	// 	log.Println(err)
	// }

	// it.Seek([]byte("t1-"))

	// for ; it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()

	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// }
	// if err := it.Err(); err != nil {
	// 	log.Println(err)
	// }

	// log.Println("--- SeekToFirst")
	// readOptions.SetPrefixSameAsStart(false) // 设为false, 否则NewIterator只输出与获取到的第一个key相同key prefix的kv
	// it= db.NewIterator(readOptions)
	// defer it.Close()

	// for it.SeekToFirst(); it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()
	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// }

	// log.Println("--- get start key")
	// it = db.NewIterator(readOptions)
	// defer it.Close()

	// it.Seek([]byte("t1-003"))

	// for ; it.Valid(); it.Next() {
	// 	// found
	// 	key := it.Key()
	// 	value := it.Value()

	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// 	break
	// }
	// if err := it.Err(); err != nil {
	// 	log.Println(err)
	// }

	// log.Println("--- get end key")
	// it = db.NewIterator(readOptions)
	// defer it.Close()

	// it.SeekForPrev([]byte("t1-030"))
	// if !it.Valid(){ // not found
	// 	it.SeekToLast()
	// }

	// for ; it.Valid(); it.Prev() {
	// 	// found
	// 	key := it.Key()
	// 	value := it.Value()

	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// 	break
	// }
	// if err := it.Err(); err != nil {
	// 	log.Println(err)
	// }

	// log.Println("--- range delete")
	// it= db.NewIterator(readOptions)
	// defer it.Close()

	// for it.SeekToFirst(); it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()
	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// }

	// wb := grocksdb.NewWriteBatch()
	// defer wb.Destroy()

	// dStartKey:=[]byte("t2-005")
	// dEndKey:=[]byte("t2-010")
	// wb.DeleteRange(dStartKey, dEndKey)

	// db.Write(writeOptions, wb)

	// it= db.NewIterator(readOptions)
	// defer it.Close()

	// for it.SeekToFirst(); it.Valid(); it.Next() {
	// 	key := it.Key()
	// 	value := it.Value()
	// 	log.Printf("Key: %v Value: %v\n", string(key.Data()), string(value.Data()))

	// 	key.Free()
	// 	value.Free()
	// }

	select {}
}

// opendb
func OpenDB() (*grocksdb.DB, error) {
	options := grocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)

	bloomFilter := grocksdb.NewBloomFilter(10)

	// rateLimiter := grocksdb.NewRateLimiter(128<<20, 10000, 10)
	// options.SetRateLimiter(rateLimiter)
	//options.EnableStatistics()
	options.SetWriteBufferSize(128 << 20)
	options.SetMaxWriteBufferNumber(3)
	options.SetMaxBackgroundCompactions(10)
	options.SetCompression(grocksdb.LZ4Compression)
	options.SetCompactionStyle(grocksdb.UniversalCompactionStyle)

	options.SetHashSkipListRep(2000000, 4, 4)

	blockBasedTableOptions := grocksdb.NewDefaultBlockBasedTableOptions()
	blockBasedTableOptions.SetBlockCache(grocksdb.NewLRUCache(64 << 20))
	blockBasedTableOptions.SetFilterPolicy(bloomFilter)
	blockBasedTableOptions.SetBlockSizeDeviation(5)
	blockBasedTableOptions.SetBlockRestartInterval(10)
	blockBasedTableOptions.SetBlockCacheCompressed(grocksdb.NewLRUCache(64 << 20))
	blockBasedTableOptions.SetCacheIndexAndFilterBlocks(true)
	blockBasedTableOptions.SetIndexType(grocksdb.KHashSearchIndexType)

	options.SetBlockBasedTableFactory(blockBasedTableOptions)
	options.SetPrefixExtractor(grocksdb.NewFixedPrefixTransform(3)) // 要通过前缀检索，需要首先创建一个前缀提取器. 注意前缀的长度, 要与Put时一致

	options.SetAllowConcurrentMemtableWrites(false)

	options.SetKeepLogFileNum(2)
	options.SetMaxLogFileSize(32 << 20)

	db, err := grocksdb.OpenDb(options, DB_PATH)

	if err != nil {
		log.Fatalln("OPEN DB error", db, err)
		db.Close()
		return nil, errors.New("fail to open db")
	} else {
		log.Println("OPEN DB success", db)
	}

	// go func(){
	// 	for {
	// 	s:=options.GetStatisticsString()
	// 		fmt.Println(s)

	// 	time.Sleep(5 * time.Second)
	// 	}
	// }()

	return db, nil
}
