package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func main() {
	OpXlsx()
}

func OpXlsx() {
	p := "."

	files, err := filepath.Glob(filepath.Join(p, "*.xlsx"))
	checkErr(err)

	for _, fn := range files {
		//parseXlsx(fn)
		injectXlsx(fn)
	}
}

func parseXlsx(fn string) {
	fmt.Printf("file:%s\n", fn)

	// excelize已默认使用charset.NewReaderLabel转其他编码到utf-8
	f, err := excelize.OpenFile(fn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var sheetName string
	for _, sname := range f.GetSheetList() {
		if sname != "" {
			sheetName = sname
			break
		}
	}

	if sheetName == "" {
		panic("no sheet")
	}
	fmt.Printf("sheet:%s\n", sheetName)

	headerRowNum, headers, err := probeHeaders(f, sheetName, 2, defaultFilterHeaders)
	checkErr(err)
	if len(headers) == 0 {
		panic("no headers")
	}

	fmt.Printf("headers:%d, %v\n", headerRowNum, strings.Join(headers, ","))

	datas, err := parseContext(f, sheetName, headerRowNum, headers)
	checkErr(err)

	for _, data := range datas {
		fmt.Println(data)
	}
}

type RowData struct {
	Time  time.Time
	Class string
	Value float64
}

func parseContext(f *excelize.File, sheetName string, startRow int, headers []string) ([]*RowData, error) {
	//like GetRows()
	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, cur := make([]*RowData, 0, 48), 0
	for rows.Next() {
		cur++
		if cur <= startRow {
			continue
		}

		row, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		if len(row) <= 1 { // need time
			continue
		}

		if len(row) < len(headers) {
			return nil, errors.New("cells less than headers")
		}

		fmt.Println("row:", row)

		baseTimeStr := row[0]
		_, err = parseDate(baseTimeStr, "2006-01-02")
		if err != nil {
			return nil, err
		}
		class := row[2]

		for i, v := range row {
			if i == 0 || i == 1 || i == 2 {
				continue
			}

			data := &RowData{
				Class: class,
			}

			data.Time, err = parseDate(baseTimeStr+" "+headers[i], "2006-01-02 15:04")
			if err != nil {
				return nil, err
			}

			data.Value, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, err
			}

			results = append(results, data)
		}
	}
	return results, nil
}

func parseDate(s string, layout string) (time.Time, error) {
	i := strings.Index(s, "\uFF1A") // 全角冒号，Unicode: U+FF1A，占用字节数: 3
	if i != -1 {
		s = s[i+3:]
	}

	s = strings.TrimSpace(s)

	var diff time.Duration
	if strings.HasSuffix(s, "24:00") {
		s = strings.ReplaceAll(s, "24:00", "23:59")
		diff = time.Minute
	}

	t, err := time.ParseInLocation(layout, s, time.Local)
	if err != nil {
		return time.Time{}, err
	}

	t = t.Add(diff)

	return t, nil
}

func defaultFilterHeaders(headers []string) bool {
	if len(headers) == 0 {
		return false
	}

	if strings.Contains(strings.Join(headers, ","), "数据单位") {
		return false
	}

	return len(headers) >= 49 && strings.LastIndex(strings.Join(headers, ","), ":") != -1
}

// headerRowNum: headers所在行, 从1开始
func probeHeaders(f *excelize.File, sheetName string, maxRow int, filterHeaders func(cols []string) bool) (int, []string, error) {
	var headerRowNum int

	rows, err := f.Rows(sheetName)
	if err != nil {
		return headerRowNum, nil, err
	}
	defer rows.Close()

	var mayHeaders []string
	for rows.Next() && maxRow > 0 {
		mayHeaders, err = rows.Columns()
		if err != nil {
			return headerRowNum, nil, err
		}

		maxRow--
		headerRowNum++

		if !filterHeaders(mayHeaders) {
			continue
		} else {
			break
		}
	}

	return headerRowNum, mayHeaders, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func injectXlsx(fn string) {
	fmt.Printf("file:%s\n", fn)

	baseName := filepath.Base(fn)
	ext := filepath.Ext(baseName)
	newXlsxName := baseName + "_err" + ext

	// excelize已默认使用charset.NewReaderLabel转其他编码到utf-8
	f, err := excelize.OpenFile(fn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var sheetName string
	for _, sname := range f.GetSheetList() {
		if sname != "" {
			sheetName = sname
			break
		}
	}

	if sheetName == "" {
		panic("no sheet")
	}
	fmt.Printf("sheet:%s\n", sheetName)

	newXlsx := excelize.NewFile()
	checkErr(err)
	defer func() {
		//Close the spreadsheet.
		if err := newXlsx.SaveAs(newXlsxName); err != nil {
			fmt.Println(err)
		}
		newXlsx.Close()
	}()
	// _, err = newXlsx.NewSheet(sheetName)
	// checkErr(err)
	newSheetName := newXlsx.GetSheetName(0) //  有默认sheet

	rows, err := f.Rows(sheetName)
	checkErr(err)
	defer rows.Close()

	s := &SaveErrorXlsx{
		f:         newXlsx,
		sheetName: newSheetName,
		cur:       0,
	}

	cur := 0
	for rows.Next() {
		cur++
		row, err := rows.Columns()
		checkErr(err)

		row = append(row, "test")

		s.InsertCols(row)
	}
}

type SaveErrorXlsx struct {
	f         *excelize.File
	sheetName string
	cur       int
}

func (s *SaveErrorXlsx) InsertCols(cols []string) {
	fmt.Println("new cols:", cols)

	nCols := make([]any, 0, len(cols))
	for _, c := range cols {
		nCols = append(nCols, c)
	}

	s.cur++
	fmt.Println("cur:", s.cur)
	err := s.f.SetSheetRow(s.sheetName, fmt.Sprintf("A%d", s.cur), &nCols)
	checkErr(err)
}
