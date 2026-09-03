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

	scanRunRepo := NewScanRunRepository()
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	scanRun, err := scanRunRepo.Create(context.Background(), tx, &domain.ScanRun{TargetNetwork: target})
	if err != nil {
		t.Fatalf("Create scan run failed: %v", err)
	}

	repo := NewMemberAttendanceRepository()

	// Create a new member attendance
	attendance := &domain.MemberAttendance{MemberID: member.ID, ScanRunID: scanRun.ID, IsPresent: true}
	created, err := repo.Create(context.Background(), tx, attendance)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.MemberID != attendance.MemberID || created.ScanRunID != attendance.ScanRunID || created.IsPresent != attendance.IsPresent {
		t.Fatalf("Created member attendance does not match")
	}

	if _, err := tx.Exec("SAVEPOINT duplicate_member_attendance"); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Create(context.Background(), tx, &domain.MemberAttendance{MemberID: member.ID, ScanRunID: scanRun.ID, IsPresent: false})
	if err == nil {
		t.Fatal("duplicate scan run/member attendance was accepted")
	}
	if _, rollbackErr := tx.Exec("ROLLBACK TO SAVEPOINT duplicate_member_attendance"); rollbackErr != nil {
		t.Fatal(rollbackErr)
	}
	if _, releaseErr := tx.Exec("RELEASE SAVEPOINT duplicate_member_attendance"); releaseErr != nil {
		t.Fatal(releaseErr)
	}

	// Get the member attendance by ID
	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || retrieved.MemberID != created.MemberID || retrieved.ScanRunID != created.ScanRunID || retrieved.IsPresent != created.IsPresent {
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
	if updated.MemberID != member.ID || updated.ScanRunID != scanRun.ID || updated.IsPresent || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated member attendance does not match expected values")
	}

	// List member attendances
	attendances, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(attendances) != 1 || attendances[0].ID != created.ID || attendances[0].ScanRunID != scanRun.ID || attendances[0].IsPresent {
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

func TestMemberAttendanceRepositoryCreateMany(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	member, err := NewMemberRepository().Create(context.Background(), tx, &domain.Member{Name: "John Doe"})
	if err != nil {
		t.Fatalf("Create member failed: %v", err)
	}

	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	secondMember, err := NewMemberRepository().Create(context.Background(), tx, &domain.Member{Name: "Jane Doe"})
	if err != nil {
		t.Fatal(err)
	}
	scanRun, err := NewScanRunRepository().Create(context.Background(), tx, &domain.ScanRun{TargetNetwork: target})
	if err != nil {
		t.Fatalf("Create scan run failed: %v", err)
	}

	attendances := []*domain.MemberAttendance{
		{MemberID: member.ID, ScanRunID: scanRun.ID, IsPresent: true},
		{MemberID: secondMember.ID, ScanRunID: scanRun.ID, IsPresent: false},
	}
	created, err := NewMemberAttendanceRepository().CreateMany(context.Background(), tx, attendances)
	if err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	if len(created) != len(attendances) {
		t.Fatalf("CreateMany returned %d attendances, want %d", len(created), len(attendances))
	}
	for i, attendance := range created {
		if attendance.ID == 0 || attendance.MemberID != attendances[i].MemberID || attendance.ScanRunID != attendances[i].ScanRunID || attendance.IsPresent != attendances[i].IsPresent || attendance.CreatedAt.IsZero() || attendance.UpdatedAt.IsZero() {
			t.Fatalf("created member attendance does not match at index %d", i)
		}
	}

	empty, err := NewMemberAttendanceRepository().CreateMany(context.Background(), tx, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("CreateMany with no attendances returned %v, %v", empty, err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
