package main

import (
	"fmt"
	"github.com/ncruces/zenity"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func lesson(rows [][]string, i int, j int, outc chan string, theseStrings []string) {
	var calendar func(int) string = func(row int) string {
		result := ""
		if row%2 != 1 {
			result = "II"
		} else {
			result = "I"
		}
		result += ";" + [6]string{"пн", "вт", "ср", "чт", "пт", "сб"}[row/14]
		result += ";" + fmt.Sprintf("%d", ((row%14)/2+1))
		return result
	}
	var ifSearched func(string, []string) string = func(record string, theseStrings []string) string {
		for _, arg := range theseStrings {
			if strings.Contains(record, arg) {
				return record
			}
		}
		return ""
	}

	if i+3 > len(rows[j]) || rows[j][i] == "" {
		outc <- ""
		return
	}

	result := fmt.Sprintf("%s", calendar(j-2))
	for i_i := i; i_i < i+4; i_i++ {
		// TODO: idea - pairs with subgroups are multilined, they ALWAYS have \n.
		corrected := strings.Replace(strings.Replace(strings.Replace(rows[j][i_i], "\t", " ", -1), "\n", " ", -1), "  ", " ", -1)
		result += fmt.Sprintf(";%s", corrected)
		if i_i == i {
			result += fmt.Sprintf(";%s", rows[1][i])
		}
	}
	result += "\n"
	outc <- ifSearched(result, theseStrings)
}

func makeTable(filename string, theseStrings []string) string {

	f, err := excelize.OpenFile(filename)
	if err != nil {
		log.Fatalf(err.Error())
		return ""
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			log.Fatalf(err.Error())
			return
		}
	}()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Расписание занятий по неделям")
	if err != nil {
		//log.Fatalf(err.Error()) # TODO мага
		return ""
	}

	outstr := ""
	for i, cell := range rows[2] {
		if cell == "Дисциплина" && i+1 < len(rows[1]) {
			var chans []chan string
			for j := 3; j < 86; j++ {
				chans = append(chans, make(chan string))
				go lesson(rows, i, j, chans[j-3], theseStrings)
			}
			for _, c := range chans {
				outstr += <-c
			}
		}
	}
	return outstr
}

func fetchTable(url string, theseStrings []string, outc chan string) {
	table, err := http.Get(url)
	if err != nil {
		outc <- ""
		log.Fatalf("Cannot reach URL: %s", url)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			outc <- ""
			return
		}
	}(table.Body)
	if strings.LastIndex(url, "/") != -1 {
		fname := url[strings.LastIndex(url, "/")+1:]
		out, err := os.Create(fname)
		if err != nil {
			outc <- ""
			log.Fatalf("Cannot create a file: %s", fname)
		}
		defer out.Close()
		if _, err := io.Copy(out, table.Body); err != nil {
			outc <- ""
			log.Fatalf("Cannot write to file: %s", fname)
		}

		outc <- makeTable(fname, theseStrings)
	}
	outc <- ""
	log.Fatalf("Crazy URL error")
}

func findRecords(theseStrings []string) string {

	resp, err := http.Get("https://mirea.ru/schedule")
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Cannot reach MIREA Schedule main page: https://mirea.ru/schedule. Code: ", resp.StatusCode)
		return ""
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	re := regexp.MustCompile(`https://webservices.mirea.ru[^\"\']*(II[TI]|IRI[^\"\']*2[^\"\']*kurs)[^\"\']*.xlsx`)
	var chans []chan string
	for i, url := range re.FindAllString(bodyString, -1) {
		chans = append(chans, make(chan string))
		go fetchTable(url, theseStrings, chans[i])
	}

	totalString := ""
	for _, c := range chans {
		totalString += <-c
	}
	return totalString

}

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

func main() {
	const defaultPath = ``
	str, err := zenity.Entry("Введите поисковый запрос:",
		zenity.Title("Расписание"))
	if err != nil {
		return
	}
	prompts := strings.Split(str, "~")

	filename, err := zenity.SelectFileSave(
		zenity.ConfirmOverwrite(),
		zenity.Filename(str + ".html"),
		zenity.FileFilters{
			{"Веб-страница HTML", []string{"*.html"}, true},
			{"Таблица CSV", []string{"*.csv"}, true},
		})


	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Unable to create file: ", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal("Unable to close file:", err.Error())
		}
	}(f)
	str = ""
	if strings.Contains(filename, ".htm") {
		str = csv2html(filename, findRecords(prompts))
	} else {
		str = findRecords(prompts)
	}
	_, err = f.WriteString(str)

	if err != nil {
		log.Fatal("Unable to write into file:", err)
	}
}
