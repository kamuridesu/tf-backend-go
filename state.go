package main

import "fmt"

type State struct {
	name     string
	contents string
	locked   bool
	db       *Database
}

func NewState(name string, db *Database) *State {
	return &State{
		name:     name,
		contents: "",
		locked:   false,
		db:       db,
	}
}

func GetAllStates() ([]*State, error) {
	return DB.GetAllStates()
}

func (s *State) Lock() error {
	if s.locked {
		return fmt.Errorf("state is locked")
	}
	s.locked = true
	s.db.UpdateState(s)
	return nil
}

func (s *State) Unlock() error {
	if !s.locked {
		return fmt.Errorf("state is not locked")
	}
	s.locked = false
	s.db.UpdateState(s)
	return nil
}

func (s *State) Update(contents string) error {
	s.contents = contents
	s.db.UpdateState(s)
	return nil
}
