package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	CheckAndCreateDB()

	nextDate, _ := NextDate(time.Now(), "20250311", "d 1")

	fmt.Println(nextDate)

	envPort := os.Getenv("TODO_PORT")

	var port string
	if envPort == "" {
		port = "7540"
	} else {
		port = envPort
	}

	fmt.Println("Запускаем сервер localhost:" + port)

	http.Handle("/", http.FileServer(http.Dir("../web")))
	http.Handle("/api/nextdate", http.HandlerFunc(NedxDateEndpoint))
	http.Handle("/api/task", http.HandlerFunc(TaskEndpoint))
	http.Handle("/api/tasks", http.HandlerFunc(GetTasksEndpoint))
	http.Handle("/api/task/done", http.HandlerFunc(DoneEndpoint))

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
