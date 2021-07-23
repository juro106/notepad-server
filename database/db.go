package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func DbInit() (*sql.DB, error) {
	var err error
	d, err := sql.Open("mysql", "user:pass@tcp(db:3306)/note?charset=utf8mb4&parseTime&loc=Local")
	if err != nil {
		log.Printf("%s\n", err)
		log.Fatal(err)
	}

	if err := d.Ping(); err != nil {
		log.Fatal("PingError: ", err)
		defer db.Close()
	}
	fmt.Printf("======>connected")
	db = d
	return db, err
}

func DbClose() {
	if db != nil {
		db.Close()
	}
}

func DbConn() *sql.DB {
	return db
}
