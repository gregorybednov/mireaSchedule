package main

import (
	"strings"
	"fmt"
	"github.com/ncruces/zenity"
	"os"
	"log"
)

func cli() {
	prompts := strings.Split("", "~")
	records := findRecords(prompts)
	fmt.Print(records)
}


func gui() {
	const defaultPath = ``
	str, err := zenity.Entry("Введите поисковый запрос:",
		zenity.Title("Расписание"))
	if err != nil {
		return
	}
	prompts := strings.Split(str, "~")
	dlg, err := zenity.Progress(
		zenity.Title("Loading..."),
		zenity.Pulsate())
	if err != nil {
		return
	}
	defer dlg.Close()

	dlg.Text("Загружаемся...")

	records := findRecords(prompts)

	dlg.Complete()

	if records == "" {
		zenity.Warning("По Вашему запросу результатов не найдено",
		zenity.Title("Таблицы скачаны, но по запросу нет результатов"),
		zenity.WarningIcon)
		os.Exit(1)
	}

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
	str = records
	if strings.Contains(filename, ".htm") {
		str = csv2html(filename, str)
	}
	_, err = f.WriteString(str)
	if err != nil {
		log.Fatal("Unable to write into file:", err)
	}
}


