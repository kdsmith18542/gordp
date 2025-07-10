package mcs

import (
	"crypto/x509"
	"io"
)

// CertBlob
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpele/ad3d569f-9f38-4a33-ae41-071b55885376
type CertBlob struct {
	CbCert uint32 // This field specifies the number of bytes in abCert.
	AbCert []byte // binary data of a single X.509 certificate.
}

// X509CertificateChain
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpele/bf2cc9cc-2b01-442e-a288-6ddfa3b80d59
type X509CertificateChain struct {
	NumCertBlobs  uint32     // Must between 2 and 200
	CertBlobArray []CertBlob // An array of CertBlob structures
	Padding       []byte     // A byte array of the length 8 + 4*NumCertBlobs is appended at the end the packet.
}

func (p *X509CertificateChain) GetPublicKey() (uint32, []byte) {
	if len(p.CertBlobArray) == 0 {
		return 0, nil
	}
	cert, err := x509.ParseCertificate(p.CertBlobArray[0].AbCert)
	if err != nil {
		return 0, nil
	}
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return 0, nil
	}
	return uint32(len(pubKeyBytes)), pubKeyBytes
}

func (p *X509CertificateChain) Verify() bool {
	if len(p.CertBlobArray) < 2 {
		return false
	}
	roots := x509.NewCertPool()
	for i, blob := range p.CertBlobArray {
		cert, err := x509.ParseCertificate(blob.AbCert)
		if err != nil {
			return false
		}
		if i == len(p.CertBlobArray)-1 {
			roots.AddCert(cert)
		}
	}
	leaf, err := x509.ParseCertificate(p.CertBlobArray[0].AbCert)
	if err != nil {
		return false
	}
	opts := x509.VerifyOptions{Roots: roots}
	_, err = leaf.Verify(opts)
	return err == nil
}

func (p *X509CertificateChain) Read(r io.Reader) {
	// TODO: Implement reading from io.Reader if needed
}
