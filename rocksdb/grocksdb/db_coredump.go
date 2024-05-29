package main

import (
	"os"

	"github.com/linxGnu/grocksdb"
)

func main() {
	dir, err := os.MkdirTemp(".", "t-")
	CheckErr(err)

	options := grocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)
	options.SetPrefixExtractor(grocksdb.NewNoopPrefixTransform())

	db, err := grocksdb.OpenDb(options, dir)
	CheckErr(err)

	db.Close()
	options.Destroy() // coredump https://github.com/facebook/rocksdb/issues/1095, 注释该行, 不销毁即可
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
