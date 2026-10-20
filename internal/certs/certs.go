package certs

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/laragal/laragal/internal/winutil"
)

func ReadCertFingerprint(certPath string) (string, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("no PEM certificate data found in %s", certPath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	hash := sha1.Sum(cert.Raw)
	return strings.ToUpper(hex.EncodeToString(hash[:])), nil
}

func BuildTrustCertCommand(certPath string) []string {
	return []string{"certutil", "-addstore", "-f", "Root", certPath}
}

func IsCertTrustedInWindowsStore(fingerprint string) (bool, error) {
	cmd := winutil.Command("certutil", "-verifystore", "Root", fingerprint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil
	}
	return strings.Contains(string(out), "Certificate is valid"), nil
}
