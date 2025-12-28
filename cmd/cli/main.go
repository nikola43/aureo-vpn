package main

import (
	"github.com/nikola43/aureo-vpn/cmd/cli/cmd"
)

// Version information (set by ldflags during build)
var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	cmd.Version = version
	cmd.BuildTime = buildTime
	cmd.Execute()
}
