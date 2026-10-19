package certs_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/laragal/laragal/app/internal/certs"
)

func TestReadCertFingerprint(t *testing.T) {
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "root.crt")

	// Generate dummy self-signed CA cert
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "LaraGal Local CA",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24),
		IsCA:      true,
	}
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	os.WriteFile(certFile, certPem, 0644)

	fingerprint, err := certs.ReadCertFingerprint(certFile)
	if err != nil {
		t.Fatalf("failed to calculate fingerprint: %v", err)
	}
	if len(fingerprint) != 40 {
		t.Fatalf("expected SHA-1 hex fingerprint length 40, got %d (%s)", len(fingerprint), fingerprint)
	}

	cmd := certs.BuildTrustCertCommand(certFile)
	if len(cmd) == 0 || cmd[0] != "certutil" {
		t.Fatalf("expected certutil command, got %v", cmd)
	}
}
