package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func OpenPostgresDB(parameters string) (*sql.DB, error) {
	db, err := sql.Open("postgres", parameters)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	err = db.Ping()

	return db, err
}
