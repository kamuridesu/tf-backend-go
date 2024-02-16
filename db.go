package main

import (
	"database/sql"
	"log"
	"reflect"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseType string

type Database struct {
	db     *sql.DB
	dbType DatabaseType
}

const (
	sqlite3  DatabaseType = "sqlite3"
	postgres DatabaseType = "postgres"
)

func buildPlaceholder(dbType DatabaseType, query string) string {
	switch dbType {
	case sqlite3:
		return query
	case postgres:
		newQuery := ""
		for i := 0; i < len(query); i++ {
			if query[i] == '?' {
				newQuery += "$" + strconv.Itoa(i+1)
				continue
			}
			newQuery += string(query[i])
		}
		return newQuery
	}
	return query
}

func StartDB(dbType DatabaseType, parameters string) (*Database, error) {
	var db *sql.DB
	var err error
	switch dbType {
	case sqlite3:
		db, err = OpenSqliteDB(parameters)
		if err != nil {
			panic(err)
		}
	case postgres:
		db, err = OpenPostgresDB(parameters)
		if err != nil {
			panic(err)
		}
	}

	sqlStmt := `CREATE TABLE IF NOT EXISTS "states" (
		"name"  TEXT NOT NULL,
		"content"     TEXT NOT NULL,
		"locked" INTEGER NOT NULL,
		PRIMARY KEY("name")
	);`

	_, err = db.Exec(sqlStmt)

	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}
	return &Database{db: db, dbType: dbType}, nil
}

func (db *Database) executeQuery(query string, params ...string) error {
	tx, err := db.db.Begin()

	if err != nil {
		log.Print(err)
		return err
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		log.Print(err)
		return err
	}

	defer stmt.Close()

	var args []reflect.Value

	for _, param := range params {
		args = append(args, reflect.ValueOf(param))
	}

	execFun := reflect.ValueOf(stmt.Exec)

	result := execFun.Call(args)

	if result[1].Interface() != nil {
		err := result[1].Interface().(error)
		log.Print(err)
		return err
	}

	err = tx.Commit()

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (db *Database) Close() {
	db.db.Close()
}

func (db *Database) SaveNewState(state *State) error {
	query := buildPlaceholder(db.dbType, `INSERT INTO states (name, content, locked) VALUES (?, ?, ?)`)
	return db.executeQuery(query)
}

func (db *Database) retrieveStates(query string, args ...string) ([]*State, error) {
	var states []*State
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return states, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(args)
	if err != nil {
		return states, err
	}

	defer rows.Close()

	for rows.Next() {
		var state State
		err := rows.Scan(&state.name, &state.contents, &state.locked)
		if err != nil {
			return states, nil
		}
		states = append(states, &state)
	}

	return states, nil
}

func (db *Database) GetAllStates() ([]*State, error) {
	query := buildPlaceholder(db.dbType, `SELECT * FROM states`)
	return db.retrieveStates(query)
}

func (db *Database) GetState(name string) (*State, error) {
	query := buildPlaceholder(db.dbType, `SELECT * FROM states WHERE name = ?`)
	states, err := db.retrieveStates(query, name)
	if err != nil || len(states) == 0 {
		return nil, err
	}
	return states[0], nil
}

func (db *Database) UpdateState(state *State) error {
	query := buildPlaceholder(db.dbType, "UPDATE state SET name=?, content=?, locked=? WHERE name=?")
	lockedRepr := "0"
	if state.locked {
		lockedRepr = "1"
	}
	return db.executeQuery(query, state.name, state.contents, lockedRepr)
}
