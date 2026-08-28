package testutil

import (
	"database/sql"
	"fmt"
	"testing"
	"uuid"

	"github.com/lib/pq"
)

type TestDBFactory struct {
	Conn       *sql.DB
	BaseConfig pq.Config
	Names      []string
}

func NewTestDB() *TestDBFactory {
	return &TestDBFactory{}
}

func (f *TestDBFactory) Connect(config pq.Config) error {
	c, err := pq.NewConnectorConfig(config)
	if err != nil {
		return err
	}

	f.Conn = sql.OpenDB(c)
	f.BaseConfig = config
	return nil
}

func (f *TestDBFactory) CreateTestDatabase(t *testing.T) *TestDB {
	t.Helper()

	name := uuid.New().String()

	if f.Conn == nil {
		t.Fatal("database connection is not established")
	}

	if _, err := f.Conn.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(name))); err != nil {
		t.Fatalf("failed to create test database %s: %v", name, err)
	}

	config := f.BaseConfig.Clone()
	config.Database = name

	return &TestDB{Name: name, Config: config}
}

func (f *TestDBFactory) Close() error {
	if f.Conn != nil {
		for _, name := range f.Names {
			if _, err := f.Conn.Exec(fmt.Sprintf("DROP DATABASE %s", pq.QuoteIdentifier(name))); err != nil {
				return fmt.Errorf("failed to drop database %s: %v", name, err)
			}
		}

		return f.Conn.Close()
	}

	return nil
}

type TestDB struct {
	Name   string
	Conn   *sql.DB
	Config pq.Config
}

func (db *TestDB) Connect(t *testing.T) {
	c, err := pq.NewConnectorConfig(db.Config)
	if err != nil {
		t.Fatalf("failed to create connector for test database %s: %v", db.Name, err)
	}

	db.Conn = sql.OpenDB(c)
}

func (db *TestDB) Close() error {
	if db.Conn != nil {
		return db.Conn.Close()
	}

	return nil
}
