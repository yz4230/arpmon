package repository

import (
	"context"
	"os"
	"testing"

	"github.com/yz4230/arpmon/pkg/db"
	"github.com/yz4230/arpmon/pkg/testutil"
)

var tdbf *testutil.TestDBFactory

func TestMain(m *testing.M) {
	tdbf = testutil.NewTestDB()

	baseConfig, err := db.LoadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	factoryConfig := baseConfig.Clone()
	factoryConfig.Database = ""
	if err = tdbf.Connect(factoryConfig); err != nil {
		panic(err)
	}

	code := m.Run()

	if err := tdbf.Close(); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func applyMigrations(t *testing.T, tdb *testutil.TestDB) {
	t.Helper()

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	err = db.ApplyMigrations(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
