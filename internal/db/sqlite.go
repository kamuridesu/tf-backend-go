package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func OpenSqliteDB(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}
