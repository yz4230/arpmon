package testutil

import (
	"database/sql"
	"fmt"
	"testing"
	"uuid"

	"github.com/lib/pq"
)

type TestDBFactory struct {
	conn       *sql.DB
	baseConfig pq.Config
	names      []string
}

func NewTestDB() *TestDBFactory {
	return &TestDBFactory{}
}

func (f *TestDBFactory) Connect(config pq.Config) error {
	c, err := pq.NewConnectorConfig(config)
	if err != nil {
		return err
	}

	f.conn = sql.OpenDB(c)
	f.baseConfig = config
	return nil
}

func (f *TestDBFactory) CreateTestDatabase(t *testing.T) *TestDB {
	t.Helper()

	name := uuid.New().String()

	if f.conn == nil {
		t.Fatal("database connection is not established")
	}

	if _, err := f.conn.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(name))); err != nil {
		t.Fatalf("failed to create test database %s: %v", name, err)
	}
	f.names = append(f.names, name)

	config := f.baseConfig.Clone()
	config.Database = name

	return &TestDB{name: name, config: config}
}

func (f *TestDBFactory) Close() error {
	if f.conn != nil {
		for _, name := range f.names {
			if _, err := f.conn.Exec(fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", pq.QuoteIdentifier(name))); err != nil {
				return fmt.Errorf("failed to drop database %s: %v", name, err)
			}
		}

		return f.conn.Close()
	}

	return nil
}

type TestDB struct {
	name   string
	Conn   *sql.DB
	config pq.Config
}

func (db *TestDB) Connect(t *testing.T) {
	c, err := pq.NewConnectorConfig(db.config)
	if err != nil {
		t.Fatalf("failed to create connector for test database %s: %v", db.name, err)
	}

	db.Conn = sql.OpenDB(c)
}

func (db *TestDB) Close() error {
	if db.Conn != nil {
		return db.Conn.Close()
	}

	return nil
}
