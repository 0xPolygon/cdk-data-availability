package dataavailability

import (
	"fmt"
	"io"
	"runtime"
)

// Populated during build, don't touch!
var (
	Version   = "v0.1.0"
	GitRev    = "undefined"
	GitBranch = "undefined"
	BuildDate = "Fri, 17 Jun 1988 01:58:00 +0200"
)

// PrintVersion prints version info into the provided io.Writer.
func PrintVersion(w io.Writer) {
	fmt.Fprint(w, GetVersionInfo())
}

// GetVersionInfo returns version information as a formatted string.
func GetVersionInfo() string {
	versionInfo := fmt.Sprintf("Version:      %s\n", Version)
	versionInfo += fmt.Sprintf("Git revision: %s\n", GitRev)
	versionInfo += fmt.Sprintf("Git branch:   %s\n", GitBranch)
	versionInfo += fmt.Sprintf("Go version:   %s\n", runtime.Version())
	versionInfo += fmt.Sprintf("Built:        %s\n", BuildDate)
	versionInfo += fmt.Sprintf("OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return versionInfo
}
