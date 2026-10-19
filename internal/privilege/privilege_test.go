package privilege_test

import (
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/privilege"
)

func TestBuildElevatedCertCommand(t *testing.T) {
	certPath := "C:\\laragal\\data\\root.crt"
	cmd := privilege.BuildElevatedCertCommand(certPath)

	if len(cmd) == 0 {
		t.Fatalf("expected non-empty command")
	}

	joined := strings.Join(cmd, " ")
	if !strings.Contains(joined, "certutil") {
		t.Errorf("command does not invoke certutil: %s", joined)
	}
	if !strings.Contains(joined, certPath) {
		t.Errorf("command does not include certPath: %s", joined)
	}
}

func TestBuildElevatedHostsScript(t *testing.T) {
	domains := []string{"mysite.test", "api.test"}
	script := privilege.BuildElevatedHostsScript(domains)

	if !strings.Contains(script, "mysite.test") {
		t.Errorf("script missing domain: %s", script)
	}
	if !strings.Contains(script, "drivers/etc/hosts") && !strings.Contains(script, "drivers\\etc\\hosts") {
		t.Errorf("script missing hosts path target: %s", script)
	}
}
