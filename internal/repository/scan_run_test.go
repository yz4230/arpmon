package repository

import (
	"context"
	"net"
	"testing"

	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/testutil"
)

func TestScanRunRepository(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewScanRunRepository()
	_, target, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	scanRun := &domain.ScanRun{TargetNetwork: target}
	created, err := repo.Create(context.Background(), tx, scanRun)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	testutil.AssertAllFieldsNonZero(t, created)

	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != created.ID || retrieved.TargetNetwork.String() != scanRun.TargetNetwork.String() {
		t.Fatalf("Retrieved scan run does not match created scan run")
	}

	scanRuns, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(scanRuns) != 1 || scanRuns[0].ID != created.ID || scanRuns[0].TargetNetwork.String() != scanRun.TargetNetwork.String() {
		t.Fatalf("List returned unexpected results")
	}

	if err := repo.DeleteByID(context.Background(), tx, created.ID); err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
