package db

import (
	"context"
	"testing"
)

func TestApplySchema(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)

	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = ApplyMigrations(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
