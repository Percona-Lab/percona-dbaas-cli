package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

const (
	tlsDefaultExpireTime = time.Hour * 24 * 180
)

type SelfSignedCerts struct {
	CA          []byte
	Certificate []byte
	PKey        []byte
}

func GenerateSelfSigned(hosts []string) (*SelfSignedCerts, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.Wrap(err, "generate serial number for root")
	}

	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Percona DBaaS Tool"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(tlsDefaultExpireTime),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "generate root key")
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, errors.Wrap(err, "create root cretificate")
	}
	rootCert := &bytes.Buffer{}
	err = pem.Encode(rootCert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, errors.Wrap(err, "marshal root cretificate")
	}

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.Wrap(err, "generate serial number for client")
	}
	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Percona XtraDB Cluster"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(tlsDefaultExpireTime),
		DNSNames:              hosts,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "generate client key")
	}
	clientDerBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, &rootTemplate, &clientKey.PublicKey, rootKey)
	if err != nil {
		return nil, errors.Wrap(err, "create client cretificate")
	}

	clientCert := &bytes.Buffer{}
	err = pem.Encode(clientCert, &pem.Block{Type: "CERTIFICATE", Bytes: clientDerBytes})
	if err != nil {
		return nil, errors.Wrap(err, "marshal client cretificate")
	}

	clientKeyEnc := &bytes.Buffer{}
	err = pem.Encode(clientKeyEnc, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})
	if err != nil {
		return nil, errors.Wrap(err, "marshal client key")
	}

	return &SelfSignedCerts{
		CA:          rootCert.Bytes(),
		Certificate: clientCert.Bytes(),
		PKey:        clientKeyEnc.Bytes(),
	}, nil
}
