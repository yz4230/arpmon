package arpscan

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type Entry struct {
	MAC net.HardwareAddr
	IP  net.IP
}

func Scan(target *net.IPNet) ([]*Entry, error) {
	c := exec.Command("arp-scan", target.String(),
		"--quiet", "--plain", "--format=${ip} ${mac}")

	output, err := c.Output()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return nil, fmt.Errorf("arp-scan exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute arp-scan: %w", err)
	}

	entries := []*Entry{}
	for line := range strings.Lines(string(output)) {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		ip := net.ParseIP(fields[0])
		if ip == nil {
			continue
		}

		mac, err := net.ParseMAC(fields[1])
		if err != nil {
			continue
		}

		entries = append(entries, &Entry{IP: ip, MAC: mac})
	}

	return entries, nil
}
