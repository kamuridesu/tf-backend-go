package main

var DB *Database

func main() {
	var err error
	DB, err = StartDB("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	serve()
}
