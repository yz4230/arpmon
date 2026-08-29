package repository

import (
	"context"
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
	member, err := memberRepo.CreateMember(context.Background(), tx, &domain.Member{Name: "John Doe"})
	if err != nil {
		t.Fatalf("CreateMember failed: %v", err)
	}

	repo := NewMemberAttendanceRepository()

	// Create a new member attendance
	attendance := &domain.MemberAttendance{MemberID: member.ID, IsPresent: true}
	created, err := repo.CreateMemberAttendance(context.Background(), tx, attendance)
	if err != nil {
		t.Fatalf("CreateMemberAttendance failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.MemberID != attendance.MemberID || created.IsPresent != attendance.IsPresent {
		t.Fatalf("Created member attendance does not match")
	}

	// Get the member attendance by ID
	retrieved, err := repo.GetMemberAttendanceByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetMemberAttendanceByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || retrieved.MemberID != created.MemberID || retrieved.IsPresent != created.IsPresent {
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
	updated, err := repo.UpdateMemberAttendance(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("UpdateMemberAttendance failed: %v", err)
	}
	if updated.MemberID != member.ID || updated.IsPresent || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated member attendance does not match expected values")
	}

	// List member attendances
	attendances, err := repo.ListMemberAttendances(context.Background(), tx)
	if err != nil {
		t.Fatalf("ListMemberAttendances failed: %v", err)
	}
	if len(attendances) != 1 || attendances[0].ID != created.ID || attendances[0].IsPresent {
		t.Fatalf("ListMemberAttendances returned unexpected results")
	}

	// Delete the member attendance
	err = repo.DeleteMemberAttendanceByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("DeleteMemberAttendanceByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
