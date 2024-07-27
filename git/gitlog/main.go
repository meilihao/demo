package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/xuri/excelize/v2"
)

var (
	gitPath string
)

const (
	repoSuffix = ".git"
)

func init() {
	flag.StringVar(&gitPath, "d", "", "git repo")
}

func main() {
	flag.Parse()

	fp := filepath.Join(gitPath, repoSuffix)

	_, err := os.Stat(fp)
	CheckIfError(err)

	r, err := git.PlainOpen(fp)
	CheckIfError(err)

	ref, err := r.Head()
	CheckIfError(err)
	fmt.Println(ref.String())

	since := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	cIter, err := r.Log(&git.LogOptions{Since: &since})
	//cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	CheckIfError(err)

	f := excelize.NewFile()
	defer func() {
		err := f.Close()
		CheckIfError(err)
	}()

	ls := [][]interface{}{}

	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		if strings.HasPrefix(c.Message, "Merge branch") {
			return nil
		}

		tmp := strings.SplitN(c.Message, "\n", 2)
		if len(tmp) > 1 {
			ls = append(ls, []interface{}{c.Hash.String(), c.Author.String(), c.Author.When.Format(time.DateTime), tmp[0], tmp[1]})
		} else {
			ls = append(ls, []interface{}{c.Hash.String(), c.Author.String(), c.Author.When.Format(time.DateTime), c.Message, ""})
		}

		return nil
	})
	if err == plumbing.ErrObjectNotFound {
		fmt.Println("done")
		err = nil
	}
	CheckIfError(err)

	for idx, row := range ls {
		cell, err := excelize.CoordinatesToCellName(1, idx+1)
		CheckIfError(err)
		f.SetSheetRow("Sheet1", cell, &row)
	}
	err = f.SaveAs(fmt.Sprintf("%s.xlsx", filepath.Base(gitPath)))
	CheckIfError(err)
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
