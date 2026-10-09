// Package buildinfo has build info injected during build.
package buildinfo

import (
	"os"
	"strconv"
)

var (
	BuildTag  string = "unknown"
	GoVersion string = "unknown"
	SystemTag string = "unknown"
)

var devmode, _ = strconv.ParseBool(os.Getenv("DEVMODE"))

func DevMode() bool {
	return devmode
}
