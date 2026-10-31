package winutil

import (
	"strconv"
	"strings"
)

// ParseListeningPIDs parses output from `netstat -ano -p TCP` and extracts unique
// PIDs that are currently in LISTENING state for the exact given TCP port.
func ParseListeningPIDs(output string, port int) []int {
	if port <= 0 {
		return nil
	}

	expectedPort := strconv.Itoa(port)
	var pids []int
	seen := make(map[int]bool)

	lines := strings.Split(output, "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		// Expected format: Proto LocalAddress ForeignAddress State PID
		if len(fields) < 5 {
			continue
		}

		if !strings.EqualFold(fields[0], "TCP") {
			continue
		}

		if !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}

		localAddr := fields[1]
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon == -1 {
			continue
		}

		addrPort := localAddr[lastColon+1:]
		if addrPort != expectedPort {
			continue
		}

		pid, err := strconv.Atoi(fields[4])
		if err != nil || pid <= 0 {
			continue
		}

		if !seen[pid] {
			seen[pid] = true
			pids = append(pids, pid)
		}
	}

	return pids
}
