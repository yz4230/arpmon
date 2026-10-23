package domain

import (
	"net"
	"time"
)

type Timestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Member struct {
	Timestamp
	ID   int
	Name string
}

type MemberAttendance struct {
	Timestamp
	ID        int
	MemberID  int
	ScanRunID int
	IsPresent bool
}

type Device struct {
	Timestamp
	ID         int
	MACAddress net.HardwareAddr
}

type ScanRun struct {
	Timestamp
	ID            int
	TargetNetwork *net.IPNet
}

type DeviceObservation struct {
	Timestamp
	ID        int
	DeviceID  int
	ScanRunID int
	IPAddress net.IP
}
