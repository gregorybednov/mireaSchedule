package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
)

func useragent() (string, string) {
	return "User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
}

func getData(url string) ([]byte, error) {

	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	return io.ReadAll(r.Body)
}

func fetchTable(url string, theseStrings []string, outc chan []record, attempt int) {
	if attempt > 10 {
		outc <- nil
		log.Fatalf("Too many attempts: %d on url: %s", attempt, url)
	}

	data, err := getData(url)
	if err != nil {
		time.Sleep(1 * time.Second)
		fetchTable(url, theseStrings, outc, attempt+1)
		return
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		time.Sleep(1 * time.Second)
		fetchTable(url, theseStrings, outc, attempt+1)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf(err.Error())
			return
		}
	}()

	rows, err := f.GetRows("Расписание занятий по неделям")
	if err != nil {
		//log.Fatalf(err.Error()) # TODO мага
		outc <- nil
		return
	}

	outc <- makeTable(rows, theseStrings)
}

func concatSlice[T any](slices ...[]T) []T {
	var result []T
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func findRecords(theseStrings []string) string {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://mirea.ru/schedule", nil)
	if err != nil {
		log.Fatalf(err.Error())
	}

	req.Header.Set(useragent())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
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
	re := regexp.MustCompile(`https://webservices.mirea.ru[^\"\']*(II[TI]|IRI|ITU)[^\"\']*.xlsx`)
	var tables []chan []record
	for i, url := range re.FindAllString(bodyString, -1) {
		tables = append(tables, make(chan []record))
		go fetchTable(url, theseStrings, tables[i], 0)
	}

	//var all_lessons []record
	var allLessons []record
	for _, c := range tables {
		lessons := <-c
		allLessons = concatSlice(allLessons, lessons)
	}

	sort.SliceStable(allLessons, func(i, j int) bool {
		return allLessons[i].Index < allLessons[j].Index
	})

	totalString := ""
	for _, lesson := range allLessons {
		totalString += lesson.Str + "\n"
	}
	return totalString
}
