// go test -v -bench=BenchmarkSingle -benchtime 100000x
// 结论:
// 1. 单线程WriteBatch比单线程put快
// 2. rocksdb多线程put比单线程慢
package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	gorocksdb "github.com/tecbot/gorocksdb"
)

func BenchmarkSingle(b *testing.B) {
	if err := os.RemoveAll("gorocksdb.db"); err != nil {
		log.Fatal(err)
	}
	db, err := OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	writeOptions := gorocksdb.NewDefaultWriteOptions()
	defer writeOptions.Destroy()

	keyFormat := "pt-%023d"
	data := make([]byte, 4123)
	log.Println("----", b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = db.Put(writeOptions, []byte(fmt.Sprintf(keyFormat, i)), data); err != nil {
			log.Println(err)
		}
	}
	b.StopTimer()
	os.RemoveAll("gorocksdb.db")
}

func BenchmarkSingleWithN(b *testing.B) {
	if err := os.RemoveAll("gorocksdb.db"); err != nil {
		log.Fatal(err)
	}
	db, err := OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	writeOptions := gorocksdb.NewDefaultWriteOptions()
	defer writeOptions.Destroy()

	keyFormat := "pt-%023d"
	data := make([]byte, 4123)
	log.Println("----", b.N)

	step := 10
	var end int

	b.ResetTimer()
	for i := 0; i < b.N; i += step {
		wb := gorocksdb.NewWriteBatch()

		end = i + step
		if end > b.N {
			end = b.N
		}
		for l := i; l < end; l++ {
			wb.Put([]byte(fmt.Sprintf(keyFormat, l)), data)
		}

		if err = db.Write(writeOptions, wb); err != nil {
			log.Println(err)
		}

		wb.Destroy()
	}
	b.StopTimer()
	os.RemoveAll("gorocksdb.db")
}

func BenchmarkMulti(b *testing.B) {
	if err := os.RemoveAll("gorocksdb.db"); err != nil {
		log.Fatal(err)
	}
	db, err := OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	n := 4
	wg := sync.WaitGroup{}

	log.Println("----", b.N)
	ses := splitN(b.N, n)
	log.Println(ses)

	wg.Add(len(ses))

	b.ResetTimer()
	for _, v := range ses {
		go putMulti(db, &wg, v)
	}

	wg.Wait()

	b.StopTimer()
	os.RemoveAll("gorocksdb.db")
}

func putMulti(db *gorocksdb.DB, wg *sync.WaitGroup, se []int) {
	keyFormat := "pt-%023d"
	data := make([]byte, 4123)
	var err error

	writeOptions := gorocksdb.NewDefaultWriteOptions()
	defer writeOptions.Destroy()

	for i := se[0]; i < se[1]; i++ {
		if err = db.Put(writeOptions, []byte(fmt.Sprintf(keyFormat, i)), data); err != nil {
			log.Println(err)
		}
	}

	wg.Done()
}

func splitN(count int, n int) [][]int {
	ls := make([][]int, 0, n)

	step := count / n
	if step == 0 {
		return [][]int{[]int{0, count}}
	}

	for i := 0; i < count; i += step {
		if i+step < count {
			ls = append(ls, []int{i, i + step})
		} else {
			ls = append(ls, []int{i, count})
		}
	}

	return ls
}
