package adapters

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var DB *sql.DB

func LoadDatabase() {
	var err error

	DBURL := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	DB, err = sql.Open("mysql", DBURL)
	if err != nil {
		log.Println("Error occcured:", err)
	}

	log.Println("We are connected to the mysql database")
}

func LoadTestDatabase() {
	var err error

	DBURL := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_PORT"),
		os.Getenv("TEST_DB_NAME"),
	)

	log.Println("Error occcured:", DBURL)
	DB, err = sql.Open("mysql", DBURL)
	if err != nil {
		log.Println("Error occcured:", err)
	}

	log.Println("We are connected to the test mysql database")
}
