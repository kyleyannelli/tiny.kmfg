package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CERT_FILE = "cert.pem"
	KEY_FILE  = "key.pem"
)

// generate or get certs
func ensureTLSCertificates() error {
	if _, err := os.Stat(CERT_FILE); err == nil {
		if _, err := os.Stat(KEY_FILE); err == nil {
			log.Info().
				Str("cert", CERT_FILE).
				Str("key", KEY_FILE).
				Msg("TLS certificates already exist")
			return nil
		}
	}

	log.Info().
		Str("cert", CERT_FILE).
		Str("key", KEY_FILE).
		Msg("Generating self-signed TLS certificates...")

	return generateSelfSignedCert(CERT_FILE, KEY_FILE)
}

func generateSelfSignedCert(certFile, keyFile string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate private key")
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"tiny.kmfg"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "*.localhost"},
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create certificate")
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		log.Error().Err(err).Str("file", certFile).Msg("Failed to create cert file")
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to write certificate")
		return err
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		log.Error().Err(err).Str("file", keyFile).Msg("Failed to create key file")
		return err
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal private key")
		return err
	}

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to write private key")
		return err
	}

	log.Info().
		Str("cert", certFile).
		Str("key", keyFile).
		Msg("Successfully generated self-signed TLS certificates")

	return nil
}

func getTLSCertPaths() (string, string) {
	certFile := os.Getenv("KMFG_TINY_CERT_FILE")
	keyFile := os.Getenv("KMFG_TINY_KEY_FILE")

	if certFile == "" {
		certFile = CERT_FILE
	}
	if keyFile == "" {
		keyFile = KEY_FILE
	}

	return certFile, keyFile
}

func logContext(event *zerolog.Event, c *fiber.Ctx) *zerolog.Event {
	return event.
		Str("uri", c.Path()).
		Str("ipAddress", c.IP()).
		Any("origin", c.Context().RemoteIP()).
		Str("referer", string(c.Request().Header.Referer())).
		Str("userAgent", string(c.Request().Header.UserAgent()))
}

func ParseTrustedProxies() []string {
	envVar := os.Getenv("KMFG_TINY_TRUSTED_IPS")
	if envVar == "" {
		return []string{}
	}

	rawIPs := strings.Split(envVar, ",")
	var trustedIPs []string

	for _, rawIP := range rawIPs {
		ip := strings.TrimSpace(rawIP)
		if ip == "" {
			continue
		}

		if net.ParseIP(ip) != nil {
			trustedIPs = append(trustedIPs, ip)
		} else {
			log.Fatal().Str("ipAddress", ip).Msg("Invalid IP given, cannot parse.")
		}
	}

	return trustedIPs
}
