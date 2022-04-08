package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

func GenerateCA(name string) (certPEM, keyPEM *pem.Block, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	validFor := time.Hour * 24 * 365 * 10 // ten years
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: name,
			Country:    []string{"CN"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return
	}
	certPEM = &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
	keyPEM = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return
}

func GenerateAndSaveCA(name, keyPath, pemPath string) error {
	certPem, keyPem, err := GenerateCA(name)
	if err != nil {
		return err
	}
	certBytes := pem.EncodeToMemory(certPem)
	keyBytes := pem.EncodeToMemory(keyPem)
	if err := ioutil.WriteFile(keyPath, keyBytes, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(pemPath, certBytes, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func GenerateNameCertKey(name string, caCertPEM, caKeyPEM []byte) (certPEM, keyPEM []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	rootCa, err := x509.ParseCertificate(caCertPEM)
	if err != nil {
		return nil, nil, err
	}
	rootCaPrivate, err := x509.ParsePKCS1PrivateKey(caKeyPEM)
	if err != nil {
		return nil, nil, err
	}
	validFor := time.Hour * 24 * 365 * 10 // ten years
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: name,
			Country:    []string{"CN"},
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	template.DNSNames = append(template.DNSNames, name)
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, rootCa, &priv.PublicKey, rootCaPrivate)
	if err != nil {
		return
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return
}
