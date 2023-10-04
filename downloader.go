package main

import (
	"io/ioutil"
	"sort"
	"regexp"
	"strings"
	"os"
	"io"
	"time"
	"net/http"
	"github.com/h2non/filetype"
	"log"
)

func useragent() (string, string) {
	return "User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
}

func fetchTable(url string, theseStrings []string, outc chan []record, attempt int) {
	if attempt > 10 {
		outc <- nil
		log.Fatalf("Too many attempts: %d on url: %s", attempt, url)
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}

	req.Header.Set(useragent())
	table, err := client.Do(req)
	if err != nil {
		outc <- nil
		log.Fatalf("Cannot reach URL: %s", url)
	}
	defer table.Body.Close()
	if strings.LastIndex(url, "/") == -1 {
		outc <- nil
		log.Fatalf("Crazy URL error")
	}

	fname := url[strings.LastIndex(url, "/")+1:]
	out, err := os.Create(fname)
	if err != nil {
		outc <- nil
		log.Fatalf("Cannot create a file: %s", fname)
	}
	defer out.Close()
	if _, err := io.Copy(out, table.Body); err != nil {
		outc <- nil
		log.Fatalf("Cannot write to file: %s", fname)
	}
	
	buf, _ := ioutil.ReadFile(fname)
	kind, _ := filetype.Match(buf)
	if kind.MIME.Value != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		os.Remove(fname)
		time.Sleep(1 * time.Second)
		fetchTable(url, theseStrings, outc, attempt+1)
		return
	}

	outc <- makeTable(fname, theseStrings)
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
	re := regexp.MustCompile(`https://webservices.mirea.ru[^\"\']*(II[TI]|IRI[^\"\']*2[^\"\']*kurs)[^\"\']*.xlsx`)
	var tables []chan []record
	for i, url := range re.FindAllString(bodyString, -1) {
		tables = append(tables, make(chan []record))
		go fetchTable(url, theseStrings, tables[i], 0)
	}

	//var all_lessons []record
	var allLessons []record;
	for _, c := range tables {
		lessons := <- c
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
