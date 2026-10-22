package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/laragal/laragal/internal/winutil"
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
		if inBlock {
			continue
		}

		// Prune obsolete manual .local local dev mappings pointing to 127.0.0.1 or ::1
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && (fields[0] == "127.0.0.1" || fields[0] == "::1") && strings.HasSuffix(strings.ToLower(fields[1]), ".local") {
			continue
		}

		lines = append(lines, line)
	}

	cleaned := strings.TrimRight(strings.Join(lines, "\n"), "\r\n")

	// Filter unique non-empty domains
	seen := make(map[string]bool)
	uniqueDomains := make([]string, 0, len(domains))
	for _, d := range domains {
		domain := strings.TrimSpace(d)
		if domain != "" && !seen[domain] {
			seen[domain] = true
			uniqueDomains = append(uniqueDomains, domain)
		}
	}

	if len(uniqueDomains) == 0 {
		return cleaned + "\n"
	}

	var block strings.Builder
	block.WriteString("\n" + BeginMarker + "\n")
	for _, domain := range uniqueDomains {
		block.WriteString(fmt.Sprintf("127.0.0.1 %s\n", domain))
		block.WriteString(fmt.Sprintf("::1 %s\n", domain))
	}
	block.WriteString(EndMarker + "\n")

	return cleaned + block.String()
}

// UpdateSystemHosts writes domains to the Windows hosts file and flushes DNS without flashing windows
func UpdateSystemHosts(domains []string) error {
	data, err := os.ReadFile(SystemHostsPath)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}
	newContent := SyncHostsContent(string(data), domains)

	// If content is already identical, do not touch the file or run any commands
	if strings.TrimSpace(newContent) == strings.TrimSpace(string(data)) {
		return nil
	}

	// 1. Attempt direct write (succeeds if running with Administrator rights)
	writeErr := os.WriteFile(SystemHostsPath, []byte(newContent), 0644)
	if writeErr == nil {
		_ = winutil.Command("ipconfig", "/flushdns").Run()
		return nil
	}

	// 2. Fallback to elevated PowerShell copy if permissions are denied
	tempFile := filepath.Join(os.TempDir(), "laragal_hosts.tmp")
	if err := os.WriteFile(tempFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp hosts file: %w", err)
	}

	psScript := fmt.Sprintf(
		`Start-Process powershell -ArgumentList '-NoProfile -Command Copy-Item -Path \"%s\" -Destination \"%s\" -Force' -Verb RunAs -WindowStyle Hidden -Wait`,
		strings.ReplaceAll(tempFile, `\`, `\\`),
		strings.ReplaceAll(SystemHostsPath, `\`, `\\`),
	)
	cmd := winutil.Command("powershell", "-NoProfile", "-Command", psScript)
	if elevErr := cmd.Run(); elevErr != nil {
		return fmt.Errorf("failed to elevate hosts update: %w (direct write error: %v)", elevErr, writeErr)
	}

	_ = winutil.Command("ipconfig", "/flushdns").Run()
	_ = os.Remove(tempFile)
	return nil
}
