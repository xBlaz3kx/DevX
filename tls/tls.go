package tls

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
)

// TLS configuration with the option to enable/disable and with paths to the certificates
type TLS struct {
	// IsEnabled is the flag to enable/disable TLS
	IsEnabled bool `yaml:"enabled" json:"enabled,omitempty" mapstructure:"enabled"`

	// RootCertificatePath is the path to the root certificate
	RootCertificatePath string `yaml:"rootCaPath" json:"rootCaPath,omitempty" mapstructure:"rootCaPath"`

	// CertificatePath is the path to the certificate
	CertificatePath string `yaml:"certPath" json:"certPath,omitempty" mapstructure:"certPath"`

	// PrivateKeyPath is the path to the private key
	PrivateKeyPath string `yaml:"keyPath" json:"keyPath,omitempty" mapstructure:"keyPath"`
}

func (t *TLS) ToTlsConfig() (*tls.Config, error) {
	if !t.IsEnabled {
		return nil, errors.New("TLS is disabled")
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	// Load Root CA certificate if path provided
	if t.RootCertificatePath != "" {
		caCert, err := os.ReadFile(t.RootCertificatePath)
		if err != nil {
			return nil, err
		} else if !certPool.AppendCertsFromPEM(caCert) {
			return nil, err
		}
	}

	// Load client certificate & private key
	certificate, err := tls.LoadX509KeyPair(t.CertificatePath, t.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{certificate},
	}, nil
}
