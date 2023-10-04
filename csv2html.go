package main

import "strings"

func newhtmlrow(row string, outc chan string) {
	res := "<tr>"
	for _, col := range strings.Split(row, ";") {
		res += "<td>" + col + "</td>"
	}
	res += "</tr>"
	outc <- res
}

func csv2html (filename string, csv string) string {
	res := "<!DOCTYPE HTML><html><head><meta charset='utf-8'/><title>"+filename+"</title><meta name='viewport' content='width=device-width, initial-scale=1.0'><style>tr, td, table {border-collapse: collapse; border: 1px solid;}</style></head><body><table>"
	var chans []chan string
	for i, row := range strings.Split(csv, "\n") {
		chans = append(chans, make(chan string))
		go newhtmlrow(row, chans[i])
	}
	for _, c := range chans {
		res += <-c
	}
	res += "</table></body></html>"
	return res
}

