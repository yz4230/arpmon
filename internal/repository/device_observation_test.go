package repository

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/testutil"
)

func TestDeviceObservationRepositoryCRUD(t *testing.T) {
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
	scanRunRepo := NewScanRunRepository()
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	parentScanRun, err := scanRunRepo.Create(context.Background(), tx, &domain.ScanRun{TargetNetwork: target})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	repo := NewDeviceObservationRepository()

	// Create a new device observation
	observation := &domain.DeviceObservation{DeviceID: firstDevice.ID, ScanRunID: parentScanRun.ID, IPAddress: net.ParseIP("192.168.1.10")}
	created, err := repo.Create(context.Background(), tx, observation)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, created)
	if created.DeviceID != observation.DeviceID || created.ScanRunID != observation.ScanRunID || !created.IPAddress.Equal(observation.IPAddress) {
		t.Fatalf("Created device observation does not match")
	}

	if _, err := tx.Exec("SAVEPOINT duplicate_device_observation"); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Create(context.Background(), tx, &domain.DeviceObservation{DeviceID: firstDevice.ID, ScanRunID: parentScanRun.ID, IPAddress: net.ParseIP("192.168.1.99")})
	if err == nil {
		t.Fatal("duplicate scan run/device observation was accepted")
	}
	if _, rollbackErr := tx.Exec("ROLLBACK TO SAVEPOINT duplicate_device_observation"); rollbackErr != nil {
		t.Fatal(rollbackErr)
	}
	if _, releaseErr := tx.Exec("RELEASE SAVEPOINT duplicate_device_observation"); releaseErr != nil {
		t.Fatal(releaseErr)
	}

	// Get the device observation by ID
	retrieved, err := repo.GetByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	testutil.AssertAllFieldsNonZero(t, retrieved)
	if retrieved.ID != created.ID || retrieved.DeviceID != created.DeviceID || retrieved.ScanRunID != created.ScanRunID {
		t.Fatalf("Retrieved device observation does not match created device observation")
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Ensure UpdatedAt will be different

	tx, err = tdb.Conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// Update the device observation
	retrieved.DeviceID = secondDevice.ID
	retrieved.IPAddress = net.ParseIP("192.168.1.11")
	updated, err := repo.Update(context.Background(), tx, retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.DeviceID != secondDevice.ID || !updated.IPAddress.Equal(retrieved.IPAddress) || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("Updated device observation does not match expected values")
	}

	// List device observations
	observations, err := repo.List(context.Background(), tx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(observations) != 1 || observations[0].ID != created.ID || observations[0].DeviceID != secondDevice.ID || observations[0].ScanRunID != parentScanRun.ID {
		t.Fatalf("List returned unexpected results")
	}

	// Create multiple device observations
	secondScanRun, err := scanRunRepo.Create(context.Background(), tx, &domain.ScanRun{TargetNetwork: target})
	if err != nil {
		t.Fatalf("Create second scan run failed: %v", err)
	}
	createdMany, err := repo.CreateMany(context.Background(), tx, []*domain.DeviceObservation{
		{DeviceID: firstDevice.ID, ScanRunID: secondScanRun.ID, IPAddress: net.ParseIP("192.168.1.12")},
		{DeviceID: secondDevice.ID, ScanRunID: secondScanRun.ID, IPAddress: net.ParseIP("192.168.1.13")},
	})
	if err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}
	if len(createdMany) != 2 {
		t.Fatalf("CreateMany returned %d device observations, want 2", len(createdMany))
	}
	for i, observation := range createdMany {
		testutil.AssertAllFieldsNonZero(t, observation)
		if observation.DeviceID != []int{firstDevice.ID, secondDevice.ID}[i] || observation.ScanRunID != secondScanRun.ID || observation.IPAddress == nil {
			t.Fatalf("Created device observation does not match at index %d", i)
		}
	}

	empty, err := repo.CreateMany(context.Background(), tx, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("CreateMany with no observations returned %v, %v", empty, err)
	}

	// Delete the device observation
	err = repo.DeleteByID(context.Background(), tx, created.ID)
	if err != nil {
		t.Fatalf("DeleteByID failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Transaction commit failed: %v", err)
	}
}
