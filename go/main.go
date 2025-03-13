package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db := CheckAndCreateDB()

	envPort := os.Getenv("TODO_PORT")

	var port string
	if envPort == "" {
		port = "7540"
	} else {
		port = envPort
	}

	fmt.Println("Запускаем сервер localhost:" + port)

	handler := NewEndpointHandlersContext(db)

	RegisterRoutes(handler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
