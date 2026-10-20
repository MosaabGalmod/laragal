package privilege

import (
	"fmt"
	"strings"

	"github.com/laragal/laragal/internal/hosts"
)

func BuildElevatedCertCommand(certPath string) []string {
	return []string{
		"powershell",
		"-NoProfile",
		"-WindowStyle",
		"Hidden",
		"-Command",
		fmt.Sprintf("Start-Process certutil -ArgumentList @('-addstore', '-f', 'Root', '%s') -Verb RunAs -WindowStyle Hidden -Wait", certPath),
	}
}

func BuildElevatedHostsScript(domains []string) string {
	domainsList := strings.Join(domains, " ")
	return fmt.Sprintf(`$hosts = '%s'; $domains = @(%s);`, hosts.SystemHostsPath, domainsList)
}
