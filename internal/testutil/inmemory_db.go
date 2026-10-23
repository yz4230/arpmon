package testutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"sync"
)

// InMemoryDB is a database/sql connection backed by an in-memory driver.
// It supports transactions, but does not execute SQL statements.
type InMemoryDB struct {
	Conn  *sql.DB
	state *inMemoryDBState
}

// NewInMemoryDB creates a database/sql connection suitable for tests that
// need to exercise transaction handling without a real database.
func NewInMemoryDB() *InMemoryDB {
	state := &inMemoryDBState{}
	return &InMemoryDB{
		Conn:  sql.OpenDB(inMemoryConnector{state: state}),
		state: state,
	}
}

// Snapshot returns the number of transactions begun, committed, and rolled
// back by this database connection.
func (db *InMemoryDB) Snapshot() (beginCount, commitCount, rollbackCount int) {
	db.state.mu.Lock()
	defer db.state.mu.Unlock()
	return db.state.beginCount, db.state.commitCount, db.state.rollbackCount
}

// Close closes the underlying database/sql connection.
func (db *InMemoryDB) Close() error {
	return db.Conn.Close()
}

type inMemoryDBState struct {
	mu            sync.Mutex
	beginCount    int
	commitCount   int
	rollbackCount int
}

type inMemoryConnector struct {
	state *inMemoryDBState
}

func (c inMemoryConnector) Connect(context.Context) (driver.Conn, error) {
	return &inMemoryConn{state: c.state}, nil
}

func (inMemoryConnector) Driver() driver.Driver {
	return inMemoryDriver{}
}

type inMemoryDriver struct{}

func (inMemoryDriver) Open(string) (driver.Conn, error) {
	return &inMemoryConn{state: &inMemoryDBState{}}, nil
}

type inMemoryConn struct {
	state *inMemoryDBState
}

func (*inMemoryConn) Prepare(string) (driver.Stmt, error) {
	return nil, driver.ErrSkip
}

func (*inMemoryConn) Close() error {
	return nil
}

func (c *inMemoryConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *inMemoryConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.state.mu.Lock()
	c.state.beginCount++
	c.state.mu.Unlock()
	return &inMemoryTx{state: c.state}, nil
}

type inMemoryTx struct {
	state *inMemoryDBState
}

func (tx *inMemoryTx) Commit() error {
	tx.state.mu.Lock()
	tx.state.commitCount++
	tx.state.mu.Unlock()
	return nil
}

func (tx *inMemoryTx) Rollback() error {
	tx.state.mu.Lock()
	tx.state.rollbackCount++
	tx.state.mu.Unlock()
	return nil
}
