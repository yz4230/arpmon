package domain

import "time"

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
	IsPresent bool
}

type Device struct {
	Timestamp
	ID         int
	MacAddress string
}

type DeviceProbe struct {
	Timestamp
	ID       int
	DeviceID int
}
