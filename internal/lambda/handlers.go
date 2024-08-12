package lambda

import (
	"fmt"
	"net/http"

	"github.com/kamuridesu/tf-backend-go/internal/db"
)

func HandleLock(name string, database *db.Database) (int, error) {
	state, err := database.GetState(name)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if state == nil {
		state = db.NewState(name, database)
		state.Database.SaveNewState(state)
	} else if state.Locked {
		return http.StatusLocked, fmt.Errorf("State %s already locked", name)
	}
	state.Lock()
	return http.StatusOK, nil
}

func HandleUnlock(name string, database *db.Database) (int, error) {
	state, err := database.GetState(name)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if state == nil {
		state = db.NewState(name, database)
		state.Database.SaveNewState(state)
	} else if !state.Locked {
		return http.StatusConflict, fmt.Errorf("State %s already unlocked", name)
	}
	state.Unlock()
	return http.StatusOK, nil
}

func HandleGet(name string, database *db.Database) (int, string, error) {
	state, err := database.GetState(name)
	if err != nil {
		return http.StatusInternalServerError, "", err
	}
	if state == nil {
		return http.StatusNotFound, "", fmt.Errorf("State %s not found", name)
	}
	return http.StatusOK, state.Contents, nil
}

func HandlePost(name, content string, database *db.Database) (int, error) {
	state, err := database.GetState(name)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if state == nil {
		state = db.NewState(name, database)
	}

	state.Update(string(content))
	return http.StatusOK, nil
}
