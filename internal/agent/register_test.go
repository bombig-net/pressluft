package agent

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRegisterPersistsLoadableKeypairAndClearsToken(t *testing.T) {
	caCert, caKey := newTestCA(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/nodes/42/register" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		csr, err := ParseCSR(req.CSR)
		if err != nil {
			t.Fatalf("ParseCSR() error = %v", err)
		}
		certPEM := signClientCSR(t, caCert, caKey, csr)
		caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})
		_ = json.NewEncoder(w).Encode(RegisterResponse{Certificate: string(certPEM), CACert: string(caPEM)})
	}))
	defer server.Close()

	previousClient := registrationHTTPClient
	registrationHTTPClient = server.Client()
	t.Cleanup(func() { registrationHTTPClient = previousClient })

	dir := t.TempDir()
	configPath := filepath.Join(dir, "agent.yaml")
	cfg := &Config{
		ServerID:          42,
		ControlPlane:      server.URL,
		CertFile:          filepath.Join(dir, "agent.crt"),
		KeyFile:           filepath.Join(dir, "agent.key"),
		CACertFile:        filepath.Join(dir, "ca.crt"),
		RegistrationToken: "bootstrap-token",
	}
	if err := cfg.SaveConfig(configPath); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	if err := Register(cfg, configPath); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if _, err := LoadClientCert(cfg); err != nil {
		t.Fatalf("LoadClientCert() error = %v", err)
	}
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	if strings.Contains(string(configBytes), "bootstrap-token") {
		t.Fatalf("registration token still present in config: %s", string(configBytes))
	}
	keyBytes, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		t.Fatalf("ReadFile(key) error = %v", err)
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "PRIVATE KEY" {
		t.Fatalf("key block = %#v, want PRIVATE KEY", block)
	}
	if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		t.Fatalf("ParsePKCS8PrivateKey() error = %v", err)
	}
}

func ParseCSR(raw string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(raw))
	return x509.ParseCertificateRequest(block.Bytes)
}

func newTestCA(t *testing.T) (*x509.Certificate, ed25519.PrivateKey) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, key)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate() error = %v", err)
	}
	return cert, key
}

func signClientCSR(t *testing.T, caCert *x509.Certificate, caKey ed25519.PrivateKey, csr *x509.CertificateRequest) []byte {
	t.Helper()
	der, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      csr.Subject,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}, caCert, csr.PublicKey, caKey)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
