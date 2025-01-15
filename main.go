package main

import (
	"os"

	"github.com/kamuridesu/tf-backend-go/cmd"
	"github.com/kamuridesu/tf-backend-go/internal/db"

	"github.com/kamuridesu/tf-backend-go/internal/lambda"
	"github.com/kamuridesu/tf-backend-go/internal/server"
)

var DB db.Database

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

	if os.Getenv("ACCESS_KEY") != "" {
		lambda.Main()
	} else {
		var err error
		users, dbParams := cmd.LoadEnvVars()

		var dbType db.DatabaseTypeInt = 0
		dbArgs := "states.db"
		if dbParams != "" {
			dbType = 1
			dbArgs = dbParams
		}
		DB, err = db.NewDatabase(dbType, dbArgs)

		if err != nil {
			panic(err)
		}

		mapUsers := mergeUsers(users)

		server.Serve(mapUsers, DB)
	}
}
