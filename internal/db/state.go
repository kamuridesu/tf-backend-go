package db

import (
	"fmt"
)

type State struct {
	Name     string
	Contents string
	Locked   bool
	Database Database
}

func NewState(name string, database Database) *State {
	return &State{
		Name:     name,
		Contents: "",
		Locked:   false,
		Database: database,
	}
}

type StateDTO struct {
	Name     string `json:"name" dynamodbav:"Name"`
	Contents string `json:"contents" dynamodbav:"Contents"`
	Locked   int64  `json:"locked" dynamodbav:"Locked"`
}

func (s *State) AsDTO() StateDTO {
	locked := 0
	if s.Locked {
		locked = 1
	}
	statedto := StateDTO{
		Name:     s.Name,
		Contents: s.Contents,
		Locked:   int64(locked),
	}

	return statedto
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
	return s.Database.UpdateState(s)
}
