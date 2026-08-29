package arpscan

import (
	"net"
)

type Entry struct {
	MAC net.HardwareAddr
	IP  net.IP
}

func Scan(target *net.IPNet) ([]*Entry, error) {
	return NewARPScanner().Scan(target)
}
