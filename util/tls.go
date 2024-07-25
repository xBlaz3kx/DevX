package util

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
)

func GetTLSConfig(certificatePath string) (*tls.Config, error) {
	certPool := x509.NewCertPool()

	data, err := os.ReadFile(certificatePath)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read CA certificate")
	}

	ok := certPool.AppendCertsFromPEM(data)
	if !ok {
		return nil, errors.Wrap(err, "couldn't read CA certificate")
	}

	return &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  certPool,
	}, nil
}
