package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/testutil"
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
	device := &domain.Device{MacAddress: "001122334455"}
	created, err := repo.Create(context.Background(), tx, device)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.MacAddress != device.MacAddress {
		t.Fatalf("Created device MAC address does not match: got %s, want %s", created.MacAddress, device.MacAddress)
	}

	// Get the device by ID
	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || retrieved.MacAddress != created.MacAddress {
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
	retrieved.MacAddress = "aabbccddeeff"
	updated, err := repo.Update(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.MacAddress != "aabbccddeeff" || !updated.UpdatedAt.After(updated.CreatedAt) {
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
