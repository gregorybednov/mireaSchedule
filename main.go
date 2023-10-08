package main

import (
	//"fmt"
	//"io"
	//"github.com/xuri/excelize/v2"
	"strconv"
	//"log"
	"os"
	"strings"
)

func lesson(rows [][]string, i int, j int, outc chan string, theseStrings []string) {
	var calendar func(int) string = func(row int) string {
		result := "I"
		if row%2 != 1 {
			result += "I"
		}
		result += ";" + [6]string{"пн", "вт", "ср", "чт", "пт", "сб"}[row/14]
		result += ";" + strconv.Itoa((row%14+1)/2)
		return result
	}
	var ifSearched func(string, []string) string = func(record string, theseStrings []string) string {
		for _, arg := range theseStrings {
			if strings.Contains(record, arg) { // TODO: не только ИЛИ, но и И
				return record
			}
		}
		return ""
	}

	if i+3 > len(rows[j]) || rows[j][i] == "" {
		outc <- ""
		return
	}

	result := calendar(j-2)
	for i_i := i; i_i < i+4; i_i++ {
		// TODO: idea - pairs with subgroups are multilined, they ALWAYS have \n.
		corrected := strings.Replace(strings.Replace(strings.Replace(rows[j][i_i], "\t", " ", -1), "\n", " ", -1), "  ", " ", -1)
		result += ";" + corrected
		if i_i == i {
			result += ";" + rows[1][i]
		}
	}
	result += "\n"
	result = ifSearched(result, theseStrings)
	outc <- result
}

type record struct {
	Index int
	Str string
}

func makeTable(rows [][]string, theseStrings []string) []record {
	var lessons []record
	for i, cell := range rows[2] {
		if cell == "Дисциплина" && i+1 < len(rows[1]) {
			var chans []chan string
			for j := 3; j < 86; j++ {
				chans = append(chans, make(chan string))
				go lesson(rows, i, j, chans[j-3], theseStrings)
			}
			for j, c := range chans {
				str := <- c
				if str == "" {
					continue
				}
				var lesson record
				lesson.Index = j + (j%2) * 1000 // thousand for sort priority
				lesson.Str = str
				lessons = append(lessons, lesson)
			}
		}
	}
	
	return lessons
}


func main() {
	if len(os.Args) > 1 && os.Args[1] == "--text-mode" {
		cli()
		os.Exit(0)
	}
	gui()
}
