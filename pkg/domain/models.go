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
	ProbeID   int
	IsPresent bool
}

type Device struct {
	Timestamp
	ID         int
	MACAddress string
}

type Probe struct {
	Timestamp
	ID     int
	Target *net.IPNet
}

type DeviceProbe struct {
	Timestamp
	ID        int
	DeviceID  int
	ProbeID   int
	IsPresent bool
}
