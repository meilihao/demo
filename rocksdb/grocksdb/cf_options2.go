package main

import (
	"os"

	"github.com/linxGnu/grocksdb"
)

func main() {
	dir, err := os.MkdirTemp(".", "t-")
	CheckErr(err)

	givenNames := []string{"default", "write"}
	options := grocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)
	options.SetCreateIfMissingColumnFamilies(true)

	oOptions := options.Clone()
	{
		oOptions.SetMemTablePrefixBloomSizeRatio(0.1)
		oOptions.SetPrefixExtractor(grocksdb.NewFixedPrefixTransform(1))

		bto := grocksdb.NewDefaultBlockBasedTableOptions()
		bto.SetBlockSize(32 << 20)
		bto.SetChecksum(0x1)
		bto.SetFilterPolicy(grocksdb.NewBloomFilterFull(1))
		bto.SetCacheIndexAndFilterBlocks(true)
		bto.SetCacheIndexAndFilterBlocksWithHighPriority(true)
		oOptions.SetBlockBasedTableFactory(bto)
	}

	dOptions := options.Clone()
	{
		dOptions.SetOptimizeFiltersForHits(false)
		dOptions.SetPrefixExtractor(grocksdb.NewNoopPrefixTransform())

		bto := grocksdb.NewDefaultBlockBasedTableOptions()
		bto.SetBlockSize(32 << 20)
		bto.SetChecksum(0x1)
		bto.SetWholeKeyFiltering(false)
		bto.SetCacheIndexAndFilterBlocks(true)
		bto.SetCacheIndexAndFilterBlocksWithHighPriority(true)
		dOptions.SetBlockBasedTableFactory(bto)
	}

	db, cfh, err := grocksdb.OpenDbColumnFamilies(options, dir, givenNames, []*grocksdb.Options{dOptions, oOptions})
	CheckErr(err)

	if len(cfh) != 2 {
		panic("cfh")
	}

	cfh[0].Destroy()
	cfh[1].Destroy()

	db.Close()
	dOptions.Destroy() // rocksdb 7.10.2/8.1.1在oracle linux 7.9会core dump, 原因不明
	oOptions.Destroy()
	options.Destroy()

	{

	}
}

func WrongCall(db *grocksdb.DB) {
	rOpt := grocksdb.NewDefaultReadOptions()
	wOpt := grocksdb.NewDefaultWriteOptions()

	k := []byte("k")
	tmp, _ := db.Get(rOpt, k)
	if tmp.Exists() {
		data := tmp.Data()
		tmp.Free()

		data[0] = 'a'
		if err := db.Put(wOpt, k, data); err != nil {
			CheckErr(err)
		}
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
