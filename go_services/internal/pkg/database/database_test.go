package database

import "testing"

func TestDB_Close_NilSafe(t *testing.T) {
	var db *DB
	if err := db.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	db = &DB{}
	if err := db.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
