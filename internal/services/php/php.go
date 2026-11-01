package php

import (
	"fmt"

	"github.com/laragal/laragal/internal/runtimes"
)

type PHPVersionInfo struct {
	Version        string `json:"version"`
	Port           int    `json:"port"`
	ExecutablePath string `json:"executable_path"`
	IniPath        string `json:"ini_path"`
}

func AssignPortForVersion(version string) int {
	return runtimes.PHPPortForVersion(version)
}

func BuildFastCGIArgs(info PHPVersionInfo) []string {
	return []string{
		"-b", fmt.Sprintf("127.0.0.1:%d", info.Port),
		"-c", info.IniPath,
	}
}
