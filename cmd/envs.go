package cmd

import (
	"os"
	"strings"
)

type User struct {
	Name     string
	Password string
}

func LoadEnvVars() (*[]User, string) {
	parsedUsers := []User{}

	users := os.Getenv("AUTH_USERS")
	dbparams := os.Getenv("POSTGRES_PARAMS")
	if users != "" {
		for _, user := range strings.Split(users, ",") {
			if strings.Contains(user, ":") {
				tmp := strings.Split(user, ":")
				if len(tmp) == 2 {
					parsedUsers = append(parsedUsers,
						User{Name: tmp[0], Password: tmp[1]})
				}
			}
		}
	}

	if dbparams != "" {
		expectedParams := []string{"host", "port", "user", "password", "dbname"}
		r := true
		for _, sep := range expectedParams {
			if !(strings.Contains(dbparams, sep)) {
				r = false
				break
			}
		}

		if !r {
			dbparams = ""
		}
	}

	return &parsedUsers, dbparams
}
