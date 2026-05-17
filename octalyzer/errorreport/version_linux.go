package errorreport

import (
	"os"
	"strings"
)

func GetOSVersion() string {
	data, err := os.ReadFile("/etc/lsb-release")
	if err != nil {
		return "Unknown"
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		parts := strings.Split(l, "=")
		if parts[0] == "DISTRIB_DESCRIPTION" {
			return parts[1]
		}
	}
	return "Unknown"
}
