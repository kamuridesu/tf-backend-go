package main

import (
	"github.com/kamuridesu/tf-backend-go/cmd"
	"github.com/kamuridesu/tf-backend-go/internal/db"
	"github.com/kamuridesu/tf-backend-go/internal/server"
)

var DB *db.Database

func mergeUsers(users *[]cmd.User) map[string]string {
	m := map[string]string{}

	for _, user := range *users {
		if user.Name != "" && user.Password != "" {
			m[user.Name] = user.Password
		}
	}

	return m
}

func main() {
	var err error
	users, dbParams := cmd.LoadEnvVars()

	var dbType db.DatabaseType = "sqlite3"
	dbArgs := "states.db"
	if dbParams != "" {
		dbType = "postgres"
		dbArgs = dbParams
	}
	DB, err = db.StartDB(dbType, dbArgs)

	if err != nil {
		panic(err)
	}

	mapUsers := mergeUsers(users)

	server.Serve(mapUsers, DB)
}
