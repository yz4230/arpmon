package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/testutil"
)

func TestDeviceProbeRepositoryCRUD(t *testing.T) {
	tdb := tdbf.CreateTestDatabase(t)
	tdb.Connect(t)
	t.Cleanup(func() { tdb.Conn.Close() })

	applyMigrations(t, tdb)

	tx, err := tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	deviceRepo := NewDeviceRepository()
	firstDevice, err := deviceRepo.CreateDevice(context.Background(), tx, &domain.Device{MacAddress: "001122334455"})
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	secondDevice, err := deviceRepo.CreateDevice(context.Background(), tx, &domain.Device{MacAddress: "aabbccddeeff"})
	if err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	probeRepo := NewProbeRepository()
	parentProbe, err := probeRepo.CreateProbe(context.Background(), tx)
	if err != nil {
		t.Fatalf("CreateProbe failed: %v", err)
	}

	repo := NewDeviceProbeRepository()

	// Create a new device probe
	probe := &domain.DeviceProbe{DeviceID: firstDevice.ID, ProbeID: parentProbe.ID}
	created, err := repo.CreateDeviceProbe(context.Background(), tx, probe)
	if err != nil {
		t.Fatalf("CreateDeviceProbe failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.DeviceID != probe.DeviceID || created.ProbeID != probe.ProbeID {
		t.Fatalf("Created device probe does not match")
	}

	// Get the device probe by ID
	retrieved, err := repo.GetDeviceProbeByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetDeviceProbeByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || retrieved.DeviceID != created.DeviceID || retrieved.ProbeID != created.ProbeID {
		t.Fatalf("Retrieved device probe does not match created device probe")
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Ensure UpdatedAt will be different

	tx, err = tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// Update the device probe
	retrieved.DeviceID = secondDevice.ID
	updated, err := repo.UpdateDeviceProbe(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("UpdateDeviceProbe failed: %v", err)
	}
	if updated.DeviceID != secondDevice.ID || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated device probe does not match expected values")
	}

	// List device probes
	probes, err := repo.ListDeviceProbes(context.Background(), tx)
	if err != nil {
		t.Fatalf("ListDeviceProbes failed: %v", err)
	}
	if len(probes) != 1 || probes[0].ID != created.ID || probes[0].DeviceID != secondDevice.ID || probes[0].ProbeID != parentProbe.ID {
		t.Fatalf("ListDeviceProbes returned unexpected results")
	}

	// Delete the device probe
	err = repo.DeleteDeviceProbeByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("DeleteDeviceProbeByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
