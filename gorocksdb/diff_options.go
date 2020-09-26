// ./diff_options -base gorocksdb.db/OPTIONS-009066 -target gorocksdb.db/OPTIONS-165905
package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"gopkg.in/ini.v1"
)

// 以base为基准比较rocksdb配置
func main() {
	baseFile := flag.String("base", "", "base option")
	targetFile := flag.String("target", "", "target option")
	cfStr := flag.String("cf", "default", "CF in OPTIONS")
	flag.Parse()

	if *baseFile == "" || *targetFile == "" {
		log.Fatal("both need not empty")
	}

	cfgBase, err := ini.Load(*baseFile)
	if err != nil {
		log.Fatal("Fail to read file: ", err)
	}
	cfgTarget, err := ini.Load(*targetFile)
	if err != nil {
		log.Fatal("Fail to read file: ", err)
	}
	cfs := strings.Split(*cfStr, ",")
	if len(cfs) == 0 {
		log.Fatal("need input CF")
	}

	dbOptionsName := "DBOptions"
	dbOptionsBase := cfgBase.Section(dbOptionsName)
	dbOptionsTarget := cfgTarget.Section(dbOptionsName)
	DiffSection(dbOptionsBase, dbOptionsTarget, dbOptionsName)

	for _, v := range cfs {
		CFOptionsName := fmt.Sprintf(`CFOptions "%s"`, v)
		CFOptionsBase := cfgBase.Section(CFOptionsName)
		CFOptionsTarget := cfgTarget.Section(CFOptionsName)
		DiffSection(CFOptionsBase, CFOptionsTarget, CFOptionsName)

		TBOptionsName := fmt.Sprintf(`TableOptions/BlockBasedTable "%s"`, v)
		TBOptionsBase := cfgBase.Section(TBOptionsName)
		TBOptionsTarget := cfgTarget.Section(TBOptionsName)
		DiffSection(TBOptionsBase, TBOptionsTarget, TBOptionsName)
	}
}

func DiffSection(base, target *ini.Section, sectionName string) {
	log.Printf("\n\n----start diff setction: %s -----\n\n", sectionName)

	onlyBase := fetchOnlyKeys(target.KeyStrings(), base.KeysHash(), false)
	onlyTarget := fetchOnlyKeys(base.KeyStrings(), target.KeysHash(), false)
	printOnly("base", onlyBase, base.KeysHash())
	printOnly("target", onlyTarget, target.KeysHash())

	kvBase := base.KeysHash()
	for _, v := range onlyBase {
		delete(kvBase, v)
	}

	fmt.Println("\n--- both ---\n")

	var tmp *ini.Key
	for k, v := range kvBase {
		tmp = target.Key(k)

		if v == tmp.String() {
			fmt.Printf("%-50s :  %s\n", k, v)
		} else {
			fmt.Printf("%-50s * %s\n", k, v+" / "+tmp.String())
		}
	}

	log.Printf("\n\n----end setction: %s -----\n\n", sectionName)
}

func fetchOnlyKeys(keys []string, kv map[string]string, needSort bool) []string {
	var only []string

	for _, v := range keys {
		delete(kv, v)
	}

	for k := range kv {
		only = append(only, k)
	}

	if needSort {
		sort.Strings(only)
	}

	return only
}

func printOnly(name string, keys []string, kv map[string]string) {
	if len(keys) == 0 {
		return
	}

	fmt.Printf("\n--- only in %s---\n", name)

	for _, v := range keys {
		fmt.Printf("%-50s :  %s\n", v, kv[v])
	}
}
