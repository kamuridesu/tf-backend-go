package db

import "testing"

func TestDB_buildPlaceholder(t *testing.T) {
	query := "INSERT INTO states (name, content, locked) VALUES (?, ?, ?)"
	expected := "INSERT INTO states (name, content, locked) VALUES ($1, $2, $3)"
	got := buildPlaceholder("postgres", query)

	if got != expected {
		t.Errorf("expected %v, got %v", expected, got)
	}
}
