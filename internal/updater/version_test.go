package updater_test

import (
	"testing"

	"github.com/laragal/laragal/internal/updater"
)

func TestIsVersionNewer(t *testing.T) {
	cases := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"8.3.16", "8.3.17", true},
		{"8.4.4", "8.4.4", false},
		{"8.4.5", "8.4.4", false},
		{"v2.8.4", "v2.9.0", true},
		{"v2.8.4", "v2.8.4", false},
		{"10.11.9", "10.11.10", true},
		{"5.0.14", "5.0.14.1", true},
		{"", "8.3.17", true},
		{"8.3.17", "", false},
	}

	for _, c := range cases {
		got := updater.IsVersionNewer(c.current, c.latest)
		if got != c.expected {
			t.Errorf("IsVersionNewer(%q, %q) = %v; want %v", c.current, c.latest, got, c.expected)
		}
	}
}

func TestExtractVersion(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"PHP 8.3.17 (cli) (built: Feb 12 2025 10:00:00) ( NTS Visual C++ 2019 x64 )", "8.3.17"},
		{"v2.8.4 h1:... caddy", "2.8.4"},
		{"mysqld from 10.11.10-MariaDB, client 15.2", "10.11.10"},
		{"Redis server v=5.0.14.1 sha=00000000:0 malloc=jemalloc-5.2.1 bits=64", "5.0.14.1"},
		{"Mailpit v1.21.8", "1.21.8"},
		{"unknown format output with no numbers", ""},
	}

	for _, c := range cases {
		got := updater.ExtractVersion(c.input)
		if got != c.expected {
			t.Errorf("ExtractVersion(%q) = %q; want %q", c.input, got, c.expected)
		}
	}
}
