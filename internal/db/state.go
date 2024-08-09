package db

import "fmt"

type State struct {
	Name     string
	Contents string
	Locked   bool
	Database *Database
}

func NewState(name string, database *Database) *State {
	return &State{
		Name:     name,
		Contents: "",
		Locked:   false,
		Database: database,
	}
}

func GetAllStates(database *Database) ([]*State, error) {
	return database.GetAllStates()
}

func (s *State) Lock() error {
	if s.Locked {
		return fmt.Errorf("state is locked")
	}
	s.Locked = true
	s.Database.UpdateState(s)
	return nil
}

func (s *State) Unlock() error {
	if !s.Locked {
		return fmt.Errorf("state is not locked")
	}
	s.Locked = false
	s.Database.UpdateState(s)
	return nil
}

func (s *State) Update(contents string) error {
	s.Contents = contents
	s.Database.UpdateState(s)
	return nil
}
