package scancmd

import (
	"log/slog"
	"net"

	"github.com/spf13/cobra"
	"github.com/yz4230/arpmon/pkg/arpscan"
)

var scanFlags struct {
	target net.IPNet
}

var ScanCmd = &cobra.Command{
	Use: "scan",
	Run: func(cmd *cobra.Command, args []string) {
		entires, err := arpscan.Scan(&scanFlags.target)
		if err != nil {
			slog.Error("Failed to scan network", "error", err)
			return
		}

		for _, entry := range entires {
			slog.Info("Found device", "IP", entry.IP, "MAC", entry.MAC.String())
		}
	},
}

func init() {
	ScanCmd.Flags().IPNetVarP(&scanFlags.target, "target", "t", net.IPNet{}, "Target network to scan (CIDR notation)")
}
