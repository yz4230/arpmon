package main

import (
	"github.com/yz4230/arpmon/cmd"
	"github.com/yz4230/arpmon/pkg/uptime"
)

func main() {
	uptime.Initialize()
	cmd.Execute()
}
