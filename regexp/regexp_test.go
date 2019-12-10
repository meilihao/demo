package main

import (
	"regexp"
	"testing"
)

// FindAllStringSubmatch 与 FindStringSubmatch 的区别（一个贪婪，一个不贪婪）
// Submatch 表示匹配分组

// FindStringSubmatch 返回 []string 	分组
func TestFindStringSubmatch(t *testing.T) {
	r := regexp.MustCompile(`^(\w+)\-(\d+)$`)

	input := "raidz3-0" // zfs raidzN group
	all := r.FindStringSubmatch(input)
	t.Log(all)
}

// FindString 非贪婪模式
func TestFindString(t *testing.T) {
	r := regexp.MustCompile("fo.?")

	input := "seafood"
	all := r.FindString(input) // output: foo
	t.Log(all)
}

// FindAllString 贪婪模式
func TestFindAllString(t *testing.T) {
	r := regexp.MustCompile("a.")

	input := "paranormal"
	all := r.FindAllString(input, -1) // output: [ar an al]
	t.Log(all)
}

// FindAllStringSubmatch 贪婪模式，分组
func TestFindAllStringSubmatch(t *testing.T) {
	r := regexp.MustCompile("a(x*)b")

	input := "-axxb-ab-"
	all := r.FindAllStringSubmatch(input, -1) // output: [["axxb" "xx"] ["ab" ""]]
	t.Log(all)
}
