package x509

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	cx509 "crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/deweppro/go-errors"
)

type (
	Cert struct {
		Public  []byte
		Private []byte
	}

	Config struct {
		Organization       string
		OrganizationalUnit string
		Country            string
		Province           string
		Locality           string
		StreetAddress      string
		PostalCode         string
	}
)

func (v *Config) ToSubject() pkix.Name {
	result := pkix.Name{}

	if len(v.Country) > 0 {
		result.Country = []string{v.Country}
	}
	if len(v.Organization) > 0 {
		result.Organization = []string{v.Organization}
	}
	if len(v.OrganizationalUnit) > 0 {
		result.OrganizationalUnit = []string{v.OrganizationalUnit}
	}
	if len(v.Locality) > 0 {
		result.Locality = []string{v.Locality}
	}
	if len(v.Province) > 0 {
		result.Province = []string{v.Province}
	}
	if len(v.StreetAddress) > 0 {
		result.StreetAddress = []string{v.StreetAddress}
	}
	if len(v.PostalCode) > 0 {
		result.PostalCode = []string{v.PostalCode}
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func generate(c *Config, ttl time.Duration, sn int64, ca *Cert, cn ...string) (*Cert, error) {
	crt := &cx509.Certificate{
		SerialNumber: big.NewInt(sn),
		Subject:      c.ToSubject(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(ttl),
		ExtKeyUsage:  []cx509.ExtKeyUsage{cx509.ExtKeyUsageClientAuth, cx509.ExtKeyUsageServerAuth},
	}

	var (
		bits int
		b    []byte
	)

	if ca == nil {
		bits = 4096
		crt.IsCA = true
		crt.BasicConstraintsValid = true
		crt.KeyUsage = cx509.KeyUsageDigitalSignature | cx509.KeyUsageCertSign
		if len(cn) > 0 {
			crt.Subject.CommonName = cn[0]
		}
	} else {
		bits = 2048
		crt.KeyUsage = cx509.KeyUsageDigitalSignature
		crt.PermittedDNSDomainsCritical = true
		for i, s := range cn {
			if i == 0 {
				crt.Subject.CommonName = cn[0]
			}
			crt.DNSNames = append(crt.DNSNames, s)
		}
	}

	pk, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, errors.WrapMessage(err, "generate private key")
	}

	if ca == nil {
		b, err = cx509.CreateCertificate(rand.Reader, crt, crt, &pk.PublicKey, pk)
	} else {
		block, _ := pem.Decode(ca.Public)
		if block == nil {
			return nil, errors.New("invalid decode public CA pem ")
		}
		var caCrt *cx509.Certificate
		caCrt, err = cx509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.WrapMessage(err, "parse CA certificate")
		}

		block, _ = pem.Decode(ca.Private)
		if block == nil {
			return nil, errors.New("invalid decode private CA pem ")
		}
		var caPK *rsa.PrivateKey
		caPK, err = cx509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.WrapMessage(err, "decode CA private key")
		}

		b, err = cx509.CreateCertificate(rand.Reader, crt, caCrt, &pk.PublicKey, caPK)
	}
	if err != nil {
		return nil, errors.WrapMessage(err, "generate certificate")
	}

	var pubPEM bytes.Buffer
	if err = pem.Encode(&pubPEM, &pem.Block{Type: "CERTIFICATE", Bytes: b}); err != nil {
		return nil, errors.WrapMessage(err, "encode public pem")
	}

	var privPEM bytes.Buffer
	if err = pem.Encode(&privPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: cx509.MarshalPKCS1PrivateKey(pk)}); err != nil {
		return nil, errors.WrapMessage(err, "encode private pem")
	}

	return &Cert{
		Public:  pubPEM.Bytes(),
		Private: privPEM.Bytes(),
	}, nil
}

func NewCertCA(c *Config, ttl time.Duration, cn string) (*Cert, error) {
	return generate(c, ttl, 1, nil, cn)
}

func NewCert(c *Config, ttl time.Duration, sn int64, ca *Cert, cn ...string) (*Cert, error) {
	return generate(c, ttl, 1, ca, cn...)
}
