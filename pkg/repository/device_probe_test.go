package repository

import (
	"context"
	"net"
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
	firstDevice, err := deviceRepo.Create(context.Background(), tx, &domain.Device{MACAddress: mustParseMAC(t, "00:11:22:33:44:55")})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	secondDevice, err := deviceRepo.Create(context.Background(), tx, &domain.Device{MACAddress: mustParseMAC(t, "aa:bb:cc:dd:ee:ff")})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	probeRepo := NewProbeRepository()
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	parentProbe, err := probeRepo.Create(context.Background(), tx, &domain.Probe{Target: target})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	repo := NewDeviceProbeRepository()

	// Create a new device probe
	probe := &domain.DeviceProbe{DeviceID: firstDevice.ID, ProbeID: parentProbe.ID, IsPresent: true}
	created, err := repo.Create(context.Background(), tx, probe)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.DeviceID != probe.DeviceID || created.ProbeID != probe.ProbeID || created.IsPresent != probe.IsPresent {
		t.Fatalf("Created device probe does not match")
	}

	// Get the device probe by ID
	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
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
	retrieved.IsPresent = false
	updated, err := repo.Update(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.DeviceID != secondDevice.ID || updated.IsPresent || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated device probe does not match expected values")
	}

	// List device probes
	probes, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(probes) != 1 || probes[0].ID != created.ID || probes[0].DeviceID != secondDevice.ID || probes[0].ProbeID != parentProbe.ID || probes[0].IsPresent {
		t.Fatalf("List returned unexpected results")
	}

	// Create multiple device probes
	createdMany, err := repo.CreateMany(context.Background(), tx, []*domain.DeviceProbe{
		{DeviceID: firstDevice.ID, ProbeID: parentProbe.ID, IsPresent: true},
		{DeviceID: secondDevice.ID, ProbeID: parentProbe.ID, IsPresent: true},
	})
	if err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}
	if len(createdMany) != 2 {
		t.Fatalf("CreateMany returned %d device probes, want 2", len(createdMany))
	}
	for i, probe := range createdMany {
		testutil.AssertAllFieldsNonZero(t, probe)
		if probe.DeviceID != []int{firstDevice.ID, secondDevice.ID}[i] || probe.ProbeID != parentProbe.ID || !probe.IsPresent {
			t.Fatalf("Created device probe does not match at index %d", i)
		}
	}

	empty, err := repo.CreateMany(context.Background(), tx, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("CreateMany with no probes returned %v, %v", empty, err)
	}

	// Delete the device probe
	err = repo.DeleteByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
