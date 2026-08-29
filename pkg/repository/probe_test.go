package repository

import (
	"context"
	"net"
	"testing"

	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/testutil"
)

func TestProbeRepository(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewProbeRepository()
	_, target, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	probe := &domain.Probe{Target: target}
	created, err := repo.Create(context.Background(), tx, probe)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	testutil.AssertAllFieldsNonZero(t, created)

	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != created.ID || retrieved.Target.String() != probe.Target.String() {
		t.Fatalf("Retrieved probe does not match created probe")
	}

	probes, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(probes) != 1 || probes[0].ID != created.ID || probes[0].Target.String() != probe.Target.String() {
		t.Fatalf("List returned unexpected results")
	}

	if err := repo.DeleteByID(context.Background(), tx, created.ID); err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
