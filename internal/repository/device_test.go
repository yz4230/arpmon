package repository

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/testutil"
)

func TestDeviceRepositoryCRUD(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewDeviceRepository()

	// Create a new device
	device := &domain.Device{MACAddress: mustParseMAC(t, "00:11:22:33:44:55")}
	created, err := repo.Create(context.Background(), tx, device)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if !bytes.Equal(created.MACAddress, device.MACAddress) {
		t.Fatalf("Created device MAC address does not match: got %s, want %s", created.MACAddress.String(), device.MACAddress.String())
	}

	// Get the device by ID
	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || !bytes.Equal(retrieved.MACAddress, created.MACAddress) {
		t.Fatalf("Retrieved device does not match created device")
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Ensure UpdatedAt will be different

	tx, err = tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// Update the device
	retrieved.MACAddress = mustParseMAC(t, "aa:bb:cc:dd:ee:ff")
	updated, err := repo.Update(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.MACAddress.String() != "aa:bb:cc:dd:ee:ff" || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated device does not match expected values")
	}

	// List devices
	devices, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(devices) != 1 || devices[0].ID != created.ID {
		t.Fatalf("List returned unexpected results")
	}

	// Delete the device
	err = repo.DeleteByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}

func TestDeviceRepositoryCreateMany(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	devices := []*domain.Device{
		{MACAddress: mustParseMAC(t, "00:11:22:33:44:55")},
		{MACAddress: mustParseMAC(t, "aa:bb:cc:dd:ee:ff")},
	}
	created, err := NewDeviceRepository().CreateMany(context.Background(), tx, devices)
	if err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	if len(created) != len(devices) {
		t.Fatalf("CreateMany returned %d devices, want %d", len(created), len(devices))
	}
	for i, device := range created {
		testutil.AssertAllFieldsNonZero(t, device)
		if !bytes.Equal(device.MACAddress, devices[i].MACAddress) {
			t.Fatalf("created device at index %d has MAC address %q, want %q", i, device.MACAddress.String(), devices[i].MACAddress.String())
		}
	}

	empty, err := NewDeviceRepository().CreateMany(context.Background(), tx, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("CreateMany with no devices returned %v, %v", empty, err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}

func mustParseMAC(t *testing.T, value string) net.HardwareAddr {
	t.Helper()

	mac, err := net.ParseMAC(value)
	if err != nil {
		t.Fatalf("net.ParseMAC(%q): %v", value, err)
	}
	return mac
}
