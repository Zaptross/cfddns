package main

import (
	"fmt"
	"log/slog"
	"os"

	dhv "github.com/zaptross/godohuver"
)

func getVersion() string {
	dat, err := os.ReadFile("/etc/program-version")

	if err != nil {
		slog.Warn("Failed to read version file", "error", err)
		dat = []byte("unknown")
	}

	version := string(dat)
	versionMessage := fmt.Sprintf("Version: %s", version)

	semver, err := dhv.ExtractSemver(version)
	if err == nil {
		latest, err := dhv.GetLatestImage("zaptross/cfddns")

		if err == nil {
			if semver != latest.Tag {
				versionMessage = fmt.Sprintf("%s\nNew version available: %s", versionMessage, latest.Tag)
			}
		}
	}

	return versionMessage
}
