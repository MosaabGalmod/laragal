package hosts_test

import (
	"strings"
	"testing"

	"github.com/laragal/laragal/app/internal/hosts"
)

func TestSyncHostsContent(t *testing.T) {
	initial := "127.0.0.1 localhost\n::1 localhost\n"
	domains := []string{"mysite.test", "api.test"}

	result := hosts.SyncHostsContent(initial, domains)

	if !strings.Contains(result, "127.0.0.1 localhost") {
		t.Errorf("original hosts content removed")
	}
	if !strings.Contains(result, "127.0.0.1 mysite.test") {
		t.Errorf("missing mysite.test")
	}
	if !strings.Contains(result, "127.0.0.1 api.test") {
		t.Errorf("missing api.test")
	}
	if !strings.Contains(result, hosts.BeginMarker) || !strings.Contains(result, hosts.EndMarker) {
		t.Errorf("missing boundary markers")
	}

	// Test update with empty domains (should remove block cleanly)
	cleaned := hosts.SyncHostsContent(result, []string{})
	if strings.Contains(cleaned, "mysite.test") {
		t.Errorf("cleaned hosts still contains removed domain")
	}
}
