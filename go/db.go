package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func CheckAndCreateDB() {
	dbFilePath := "../scheduler.db"
	_, err := os.Stat(dbFilePath)

	var install bool
	if err != nil {
		install = true
	}
	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX

	if install {
		db, err := sql.Open("sqlite3", dbFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		createTable := `CREATE TABLE scheduler (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		date TEXT NOT NULL CHECK(length(date) = 8), -- Дата в формате YYYYMMDD
    		title TEXT NOT NULL,
    		comment TEXT,
    		repeat TEXT CHECK(length(repeat) <= 128)
		);

			CREATE INDEX idx_date ON scheduler(date);
		`
		_, err = db.Exec(createTable)
		if err != nil {
			log.Fatal(err)
		}
	}
}
