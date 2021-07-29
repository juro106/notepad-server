package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func DbInit() (*sql.DB, error) {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// d, err := sql.Open("mysql", "user:pass@tcp(db:3306)/note?charset=utf8mb4&parseTime&loc=Local")
	d, err := sql.Open(
		"mysql",
		os.Getenv("DB_USER")+":"+
			os.Getenv("DB_PASSWORD")+"@tcp("+
			os.Getenv("DB_HOST")+":"+
			os.Getenv("DB_PORT")+")/"+
			os.Getenv("DB_NAME")+"?"+
			os.Getenv("DB_OPTION"))
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
