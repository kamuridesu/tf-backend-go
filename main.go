package main

var DB *Database

func mergeUsers(users *[]User) map[string]string {
	var m map[string]string

	for _, user := range *users {
		m[user.name] = user.password
	}

	return m
}

func main() {
	var err error
	users, dbParams := LoadEnvVars()

	var dbType DatabaseType = "sqlite3"
	dbArgs := "states.db"
	if dbParams != "" {
		dbType = "postgres"
		dbArgs = dbParams
	}
	DB, err = StartDB(dbType, dbArgs)

	if err != nil {
		panic(err)
	}

	mapUsers := mergeUsers(users)

	serve(mapUsers)
}
