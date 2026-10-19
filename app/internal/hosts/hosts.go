package hosts

import (
	"fmt"
	"os"
	"strings"
)

const (
	BeginMarker     = "# --- BEGIN LARAGAL MANAGED DOMAINS ---"
	EndMarker       = "# --- END LARAGAL MANAGED DOMAINS ---"
	SystemHostsPath = "C:\\Windows\\System32\\drivers\\etc\\hosts"
)

func SyncHostsContent(content string, domains []string) string {
	var lines []string
	inBlock := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == BeginMarker {
			inBlock = true
			continue
		}
		if trimmed == EndMarker {
			inBlock = false
			continue
		}
		if !inBlock {
			lines = append(lines, line)
		}
	}

	cleaned := strings.TrimRight(strings.Join(lines, "\n"), "\r\n")

	if len(domains) == 0 {
		return cleaned + "\n"
	}

	var block strings.Builder
	block.WriteString("\n" + BeginMarker + "\n")
	for _, domain := range domains {
		d := strings.TrimSpace(domain)
		if d != "" {
			block.WriteString(fmt.Sprintf("127.0.0.1 %s\n", d))
		}
	}
	block.WriteString(EndMarker + "\n")

	return cleaned + block.String()
}

func UpdateSystemHosts(domains []string) error {
	data, err := os.ReadFile(SystemHostsPath)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}
	newContent := SyncHostsContent(string(data), domains)
	return os.WriteFile(SystemHostsPath, []byte(newContent), 0644)
}
