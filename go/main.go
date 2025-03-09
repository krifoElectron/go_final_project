package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	CheckAndCreateDB()

	port := os.Getenv("TODO_PORT")
	fmt.Println("Запускаем сервер localhost:" + port)
	http.Handle("/", http.FileServer(http.Dir("../web")))
	http.Handle("/api/nextdate", http.HandlerFunc(NedxDateEndpoint))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}

func NextDate(now time.Time, eventDateString string, repeat string) (string, error) {
	if repeat == "" {
		err := errors.New("emit macho dwarf: elf header corrupted")

		return "", err
	}

	eventDate, err := time.Parse("20060102", eventDateString)
	if err != nil {
		return "", err
	}

	result := strings.Split(repeat, " ")
	fmt.Println(result)

	// Базовые правила
	if len(result) == 2 || result[0] == "y" {
		if result[0] == "y" {
			nextDate := eventDate.AddDate(1, 0, 0)
			for nextDate.Before(now) {
				nextDate = nextDate.AddDate(1, 0, 0)
			}
			return nextDate.Format("20060102"), nil
		}

		if result[0] == "d" {
			period, err := strconv.Atoi(result[1])

			if period > 400 {
				err = errors.New("период должен быть не более 400 дней")
				return "", err
			}

			if err != nil {
				return "", err
			}
			nextDate := eventDate.AddDate(0, 0, period)
			for nextDate.Before(now) {
				nextDate = nextDate.AddDate(0, 0, period)
			}
			return nextDate.Format("20060102"), nil
		}
	}

	err = errors.New("неверный формат repeat")
	return "", err
}

func NedxDateEndpoint(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nowDate, err := time.Parse("20060102", now)
	if err != nil {
		w.Write([]byte("Неверный формат now"))
		return
	}

	nextDate, err := NextDate(nowDate, date, repeat)

	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(nextDate))
}
