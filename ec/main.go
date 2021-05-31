// 纠删码测试
// 将一个文件拆分成10分，删除其中的任意三分，尝试还原文件
// 这边需要注意的一点是文件的拆分后的顺序和还原的顺序是相关的,顺序错误是无法还原的

// 其中Encoder接口有以下几个关键的函数:
// Verify(shards [][]byte) (bool, error): 每个分片都是[]byte类型，分片集合就是[][]byte类型，传入所有分片，如果有任意的分片数据错误，就返回false
// Split(data []byte) ([][]byte, error): 将原始数据按照规定的分片数进行切分。注意：数据没有经过拷贝，所以修改分片也就是修改原数据
// Reconstruct(shards [][]byte) error: 这个函数会根据shards中完整的分片，重建其他损坏的分片
// Join(dst io.Writer, shards [][]byte, outSize int) error: 将shards合并成完整的原始数据并写入dst这个Writer中

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/klauspost/reedsolomon"
)

var (
	srcFile      string // 原始文件
	dstDir       string // 目标目录
	recoverName  string // 还原后的文件名称
	dataShards   int    // 数据分片
	parityShards int    // 校验分片
	op           string // 操作动作
)

func init() {
	flag.StringVar(&srcFile, "srcFile", "main.go", "原始文件名称")
	flag.StringVar(&dstDir, "dstDir", "dstDir", "目标目录")
	flag.StringVar(&recoverName, "recoverName", "recoverName", "还原后的文件名称")
	flag.StringVar(&op, "op", "", "split和recover二选一,split会将一个文件拆分成类似10个数据文件和3和校验文件,recover的时候可以删除目标目录下的三个文件做还原即可")
	flag.IntVar(&dataShards, "dataShards", 10, "数据分片个数")
	flag.IntVar(&parityShards, "parityShards", 3, "校验分片个数")
	flag.Parse()
}

// ./ec -op=split先将文件拆分
// 删除当前目录下的子文件夹dstDir里面的任意三个文件
// ./ec -op=recover -recoverName="r.go" 将文件还原
// 最后比对md5发现是一致的
func main() {
	if op == "split" {
		os.ReadDir(dstDir)
		os.MkdirAll(dstDir, 0755)

		if err := splitFile(); err != nil {
			fmt.Println(err)
			return
		}
	} else if op == "recover" {
		if err := recoverFile(); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("错误的操作动作参数,op必须为split或者recover")
		return
	}
}

// 还原文件验证
func recoverFile() error {
	if recoverName == "" {
		return fmt.Errorf("还原后的文件名称%s不能为空", recoverName)
	}

	// 数据分10片和校验3片
	enc, err := reedsolomon.New(dataShards, parityShards, reedsolomon.WithAutoGoroutines(dataShards+parityShards))
	if err != nil {
		return fmt.Errorf("创建数据分片和校验分片失败,%s", err.Error())
	}

	shards := make([][]byte, dataShards+parityShards)
	for i := range shards {
		splitName := fmt.Sprintf("%ssplit%010d", dstDir, i)
		// 不管文件是否存在，需要保留原先的顺序
		if shards[i], err = ioutil.ReadFile(filepath.Join(dstDir, splitName)); err != nil {
			fmt.Printf("读取文件[%s]失败,%s\n", splitName, err.Error())

			continue
		}
		fmt.Println(splitName)
	}

	ok, err := enc.Verify(shards)
	if ok {
		fmt.Println("非常好,数据块和校验块都完整")
	} else {
		if err = enc.Reconstruct(shards); err != nil {
			return fmt.Errorf("重建其他损坏的分片失败,%s", err.Error())
		}

		if ok, err = enc.Verify(shards); err != nil {
			return fmt.Errorf("数据块校验失败2,%s", err.Error())
		}
		if !ok {
			return fmt.Errorf("重建其他损坏的分片后数据还是不完整,文件损坏")
		}

	}
	f, err := os.Create(filepath.Join(dstDir, recoverName))
	if err != nil {
		return fmt.Errorf("创建还原文件[%s]失败,%s", recoverName, err.Error())
	}
	// shards的dataShards总大小和原先的是不是一致的
	// 因此实际生产需要一开始拆分文件时候就记录源文件的大小
	//if err = enc.Join(f, shards, len(shards[0])*dataShards); err != nil {
	_, ln, err := GetFileLenAndMd5(srcFile)
	if err != nil {
		return fmt.Errorf("计算原始文件[%s]大小失败,%s", srcFile, err.Error())
	}
	if err = enc.Join(f, shards, int(ln)); err != nil {
		return fmt.Errorf("写还原文件[%s]失败,%s", recoverFile(), err.Error())
	}
	return nil
}

// 分隔文件处理
func splitFile() error {
	// 数据分10片和校验3片
	enc, err := reedsolomon.New(dataShards, parityShards, reedsolomon.WithAutoGoroutines(dataShards+parityShards))
	if err != nil {
		return fmt.Errorf("创建数据分片和校验分片失败,%s", err.Error())
	}

	bigfile, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("读取原始文件[%s]失败,%s", srcFile, err.Error())
	}
	fmt.Printf("src: %s,%d\n", srcFile, len(bigfile))

	// 将原始数据按照规定的分片数进行切分
	shards, err := enc.Split(bigfile)
	if err != nil {
		return fmt.Errorf("针对原始文件[%s]拆分成数据[%d]块,校验[%d]块失败,%s", srcFile, dataShards, parityShards, err.Error())
	}

	// 编码校验块
	if err = enc.Encode(shards); err != nil {
		return fmt.Errorf("编码校验块失败,%s", err.Error())
	}
	sum := 0
	for i := range shards {
		splitName := fmt.Sprintf("%ssplit%010d", dstDir, i)
		sum += len(shards[i])
		fmt.Printf("%s,%d\n", splitName, len(shards[i]))
		if err = ioutil.WriteFile(filepath.Join(dstDir, splitName), shards[i], 0600); err != nil {
			return fmt.Errorf("原始文件[%s]拆分文件[%s]写失败,%s", srcFile, splitName, err.Error())
		}
	}
	fmt.Printf("shards: %ssplit,%d\n", dstDir, sum)

	return nil
}

func GetFileLenAndMd5(p string) (string, int64, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return "", 0, err
	}

	return "", fi.Size(), nil
}
