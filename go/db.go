package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func GetDB() *sql.DB {
	rootDir := GetRootDirectory()

	db, err := sql.Open("sqlite3", rootDir+"/"+DB_FILE_NAME)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}

func CheckAndCreateDB() *sql.DB {
	rootDir := GetRootDirectory()

	_, err := os.Stat(rootDir + "/" + DB_FILE_NAME)

	var install bool
	if err != nil {
		install = true
	}
	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX

	db := GetDB()

	if install {
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

	return db
}
