package repository

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/testutil"
)

func TestMemberAttendanceRepositoryCRUD(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	memberRepo := NewMemberRepository()
	member, err := memberRepo.Create(context.Background(), tx, &domain.Member{Name: "John Doe"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	probeRepo := NewProbeRepository()
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	probe, err := probeRepo.Create(context.Background(), tx, &domain.Probe{Target: target})
	if err != nil {
		t.Fatalf("Create probe failed: %v", err)
	}

	repo := NewMemberAttendanceRepository()

	// Create a new member attendance
	attendance := &domain.MemberAttendance{MemberID: member.ID, ProbeID: probe.ID, IsPresent: true}
	created, err := repo.Create(context.Background(), tx, attendance)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.MemberID != attendance.MemberID || created.ProbeID != attendance.ProbeID || created.IsPresent != attendance.IsPresent {
		t.Fatalf("Created member attendance does not match")
	}

	// Get the member attendance by ID
	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || retrieved.MemberID != created.MemberID || retrieved.ProbeID != created.ProbeID || retrieved.IsPresent != created.IsPresent {
		t.Fatalf("Retrieved member attendance does not match created member attendance")
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Ensure UpdatedAt will be different

	tx, err = tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// Update the member attendance
	retrieved.IsPresent = false
	updated, err := repo.Update(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.MemberID != member.ID || updated.ProbeID != probe.ID || updated.IsPresent || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated member attendance does not match expected values")
	}

	// List member attendances
	attendances, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(attendances) != 1 || attendances[0].ID != created.ID || attendances[0].ProbeID != probe.ID || attendances[0].IsPresent {
		t.Fatalf("List returned unexpected results")
	}

	// Delete the member attendance
	err = repo.DeleteByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
