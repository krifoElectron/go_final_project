package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("TODO_PORT")
	fmt.Println("Запускаем сервер localhost:" + port)
	http.Handle("/", http.FileServer(http.Dir("../web")))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}
