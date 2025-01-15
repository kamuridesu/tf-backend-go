package db

import (
	"database/sql"
	"errors"
	"strconv"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseTypeInt int

const (
	SQLite = iota
	PostgreSQL
	Dynamo
)

type Query struct {
	dbType DatabaseTypeInt
	query  string
}

func newQuery(dbType DatabaseTypeInt, query string) *Query {
	return &Query{dbType: dbType, query: query}
}

func (q *Query) getQuery() string {
	switch q.dbType {
	case SQLite:
		return q.query
	case PostgreSQL:
		newQuery := ""
		counter := 1
		for i := 0; i < len(q.query); i++ {
			if q.query[i] == '?' {
				newQuery += "$" + strconv.Itoa(counter)
				counter++
				continue
			}
			newQuery += string(q.query[i])
		}
		return newQuery
	}
	return q.query
}

type Queries struct {
	createTable     *Query
	insertNewState  *Query
	selectAllStates *Query
	selectState     *Query
	updateState     *Query
}

func buildQueries(dbType DatabaseTypeInt) *Queries {
	return &Queries{
		createTable: newQuery(dbType, `CREATE TABLE IF NOT EXISTS "states" (
		"name"  TEXT NOT NULL,
		"content"     TEXT NOT NULL,
		"locked" INTEGER NOT NULL,
		PRIMARY KEY("name")
	);`),
		insertNewState:  newQuery(dbType, `INSERT INTO states (name, content, locked) VALUES (?, ?, ?)`),
		selectAllStates: newQuery(dbType, `SELECT * FROM states`),
		selectState:     newQuery(dbType, `SELECT * FROM states WHERE name = ?`),
		updateState:     newQuery(dbType, "UPDATE states SET name=?, content=?, locked=? WHERE name=?"),
	}
}

type Database interface {
	Connect() error
	Disconenct() error
	SaveNewState(state *State) error
	GetState(name string) (*State, error)
	UpdateState(state *State) error
}

type GenericDB struct {
	db      *sql.DB
	queries *Queries
}

func (g *GenericDB) Connect() error {
	return g.db.Ping()
}

func (g *GenericDB) Disconenct() error {
	return g.db.Close()
}

func (g *GenericDB) createTable() error {
	_, err := g.db.Exec(g.queries.createTable.getQuery())
	return err
}

func (g *GenericDB) SaveNewState(state *State) error {
	strLocked := "0"
	if state.Locked {
		strLocked = "1"
	}
	_, err := g.db.Exec(g.queries.insertNewState.getQuery(), state.Name, state.Contents, strLocked)
	return err
}

func (g *GenericDB) GetState(name string) (*State, error) {
	var state State
	var locked int
	row := g.db.QueryRow(g.queries.selectState.getQuery(), name)
	err := row.Scan(&state.Name, &state.Contents, &locked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	state.Locked = locked == 1
	state.Database = g
	return &state, nil
}

func (g *GenericDB) UpdateState(state *State) error {
	strLocked := "0"
	if state.Locked {
		strLocked = "1"
	}
	_, err := g.db.Exec(g.queries.updateState.getQuery(), state.Name, state.Contents, strLocked, state.Name)
	return err
}

func openPostgresDB(parameters string) (*sql.DB, error) {
	db, err := sql.Open("postgres", parameters)
	if err != nil {
		return nil, err
	}

	err = db.Ping()

	return db, err
}

func openSqliteDB(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewDatabase(dbType DatabaseTypeInt, dbParams string) (Database, error) {
	var db *sql.DB
	var err error
	switch dbType {
	case SQLite:
		db, err = openSqliteDB(dbParams)
	case PostgreSQL:
		db, err = openPostgresDB(dbParams)
	case Dynamo:
		db, err := OpenDynamoDB()
		if err != nil {
			return nil, err
		}
		return db, nil
	default:
		return nil, errors.New("unknown database type")
	}
	if err != nil {
		return nil, err
	}
	queries := buildQueries(dbType)
	genericDB := &GenericDB{db: db, queries: queries}
	err = genericDB.Connect()
	if err != nil {
		return nil, err
	}
	err = genericDB.createTable()
	if err != nil {
		return nil, err
	}
	return genericDB, err
}
