package main

import "fmt"

type State struct {
	name     string
	contents string
	locked   bool
}

func NewState(name string) *State {
	return &State{
		name:     name,
		contents: "",
		locked:   false,
	}
}

func GetAllStates() []*State {
	return nil
}

func (s *State) Lock() error {
	if s.locked {
		return fmt.Errorf("state is locked")
	}
	s.locked = true
	return nil
}

func (s *State) Unlock() error {
	if !s.locked {
		return fmt.Errorf("state is not locked")
	}
	s.locked = false
	return nil
}

func (s *State) Update(contents string) error {
	s.contents = contents
	return nil
}
