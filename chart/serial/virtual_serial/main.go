package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Range struct {
	Start int
	End   int
	Count int
	Nums  []int
}

func main() {
	ls := GenerateRaw(64, 1000)

	step := 30
	start := ls[0]
	end := ls[len(ls)-1]

	if i := start % step; i != 0 {
		start -= i
	}
	if i := end % step; i != 0 {
		end += (step - i)
	}
	n := (end - start) / step

	// count nicety
	if ls[len(ls)-1] == end {
		n += 1 // last one overflow
	}
	fmt.Printf("nums: %+v\n", ls)
	fmt.Printf("start: %d, end: %d, n: %d, step: %d\n", start, end, n, step)

	vrs := make([]*Range, 0, n)
	r := &Range{
		Start: start,
		End:   start + step,
	}

	for _, v := range ls {
		if !(v >= r.Start && v < r.End) {
			for {
				vrs = append(vrs, r)

				start += step

				r = &Range{
					Start: start,
					End:   start + step,
				}

				if v >= r.Start && v < r.End {
					break
				}
			}
		}

		r.Count += 1
		r.Nums = append(r.Nums, v)
	}
	if vrs[len(vrs)-1] != r { // add last one
		vrs = append(vrs, r)
	}

	c := 0
	for i, v := range vrs {
		c += v.Count

		fmt.Printf("range %02d: %+v\n", i, v)
	}

	fmt.Printf("count: %d, want: %d\n", c, len(ls))
	fmt.Printf("range count: %d, want: %d\n", len(vrs), n)
}

func GenerateRaw(n, m int) []int {
	rand.Seed(time.Now().UnixNano())

	ls := make([]int, 0, n)
	for i := 0; i < n; i++ {
		ls = append(ls, rand.Intn(m))
	}

	if ls[len(ls)-1] != m { // 模拟边界
		ls = append(ls, m)
	}

	sort.Ints(ls)

	return ls
}
