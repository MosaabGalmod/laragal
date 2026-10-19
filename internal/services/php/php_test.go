package php_test

import (
	"testing"

	"github.com/laragal/laragal/internal/services/php"
)

func TestAssignPortForVersion(t *testing.T) {
	cases := map[string]int{
		"8.1": 9081,
		"8.2": 9082,
		"8.3": 9083,
		"8.4": 9084,
		"7.4": 9074,
	}

	for ver, expectedPort := range cases {
		port := php.AssignPortForVersion(ver)
		if port != expectedPort {
			t.Errorf("version %s: expected port %d, got %d", ver, expectedPort, port)
		}
	}
}

func TestBuildFastCGIArgs(t *testing.T) {
	info := php.PHPVersionInfo{
		Version:        "8.3",
		Port:           9083,
		ExecutablePath: "C:\\laragal\\bin\\php\\8.3\\php-cgi.exe",
		IniPath:        "C:\\laragal\\bin\\php\\8.3\\php.ini",
	}

	args := php.BuildFastCGIArgs(info)
	expected := []string{"-b", "127.0.0.1:9083", "-c", "C:\\laragal\\bin\\php\\8.3\\php.ini"}

	for i, arg := range expected {
		if args[i] != arg {
			t.Fatalf("expected arg %s, got %s", arg, args[i])
		}
	}
}
