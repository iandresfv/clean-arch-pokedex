// Command tlscert generates a self-signed certificate for local HTTPS.
//
// Enabling TLS also enables HTTP/2: Go's server negotiates it automatically
// through ALPN (Application-Layer Protocol Negotiation), a TLS extension in
// which client and server agree on the application protocol during the
// handshake. No further configuration is involved, which is why there is no
// HTTP/2 switch anywhere in this project.
//
// The certificate is self-signed and therefore only suitable for development.
// In production TLS is normally terminated at an ingress controller or reverse
// proxy; supporting it in the process keeps the deployment self-contained and
// makes that substitution a removal rather than an addition.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	hosts := flag.String("host", "localhost,127.0.0.1,::1", "comma-separated DNS names and IPs")
	outDir := flag.String("out", "certs", "directory to write cert.pem and key.pem into")
	validFor := flag.Duration("valid-for", 365*24*time.Hour, "certificate lifetime")
	flag.Parse()

	// P-256 rather than RSA: comparable security at a fraction of the key size,
	// with a much faster handshake.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return fmt.Errorf("generating serial number: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Clean Arch Pokedex (development)"},
			CommonName:   "pokedex-api",
		},
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(*validFor),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Modern clients ignore CommonName entirely and validate against the
	// Subject Alternative Name extension, so every host must appear here.
	for h := range strings.SplitSeq(*hosts, ",") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	certPath := filepath.Join(*outDir, "cert.pem")
	if err := writePEM(certPath, "CERTIFICATE", der, 0o644); err != nil {
		return err
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshalling private key: %w", err)
	}

	keyPath := filepath.Join(*outDir, "key.pem")
	// 0600: a private key readable by other users on the machine is not private.
	if err := writePEM(keyPath, "EC PRIVATE KEY", keyDER, 0o600); err != nil {
		return err
	}

	fmt.Printf("wrote %s and %s\n", certPath, keyPath)
	fmt.Printf("hosts: %v %v\n", template.DNSNames, template.IPAddresses)
	fmt.Printf("valid until: %s\n", template.NotAfter.Format(time.RFC3339))
	fmt.Println()
	fmt.Println("Enable it with:")
	fmt.Printf("  TLS_ENABLED=true TLS_CERT_PATH=%s TLS_KEY_PATH=%s make run\n", certPath, keyPath)
	return nil
}

func writePEM(path, blockType string, der []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	return nil
}
