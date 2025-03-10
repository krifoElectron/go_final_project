package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	CheckAndCreateDB()

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
	http.Handle("/api/tasks", http.HandlerFunc(GetTaskskEndpoint))

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
