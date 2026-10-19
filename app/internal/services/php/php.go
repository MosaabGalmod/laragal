package php

import (
	"fmt"
	"strconv"
	"strings"
)

type PHPVersionInfo struct {
	Version        string `json:"version"`
	Port           int    `json:"port"`
	ExecutablePath string `json:"executable_path"`
	IniPath        string `json:"ini_path"`
}

func AssignPortForVersion(version string) int {
	clean := strings.ReplaceAll(version, ".", "")
	if val, err := strconv.Atoi(clean); err == nil {
		return 9000 + val
	}
	return 9083 // Fallback default port
}

func BuildFastCGIArgs(info PHPVersionInfo) []string {
	return []string{
		"-b", fmt.Sprintf("127.0.0.1:%d", info.Port),
		"-c", info.IniPath,
	}
}
