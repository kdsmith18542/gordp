package core

import (
	"bufio"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/GoFeGroup/gordp/glog"
	"github.com/huin/asn1ber"
)

type Stream struct {
	c net.Conn
	b *bufio.ReadWriter

	r func([]byte) (int, error)
	w func([]byte) (int, error)
}

func (s *Stream) Read(b []byte) (n int, err error) {
	return s.r(b)
}

func (s *Stream) Write(b []byte) (n int, err error) {
	return s.w(b)
}

func (s *Stream) Peek(n int) []byte {
	if s.b == nil {
		s.b = bufio.NewReadWriter(bufio.NewReader(s.c), bufio.NewWriter(s.c))
		s.r = func(b []byte) (int, error) { return s.b.Read(b) }
		s.w = func(b []byte) (int, error) { return s.b.Write(b) }
	}
	d, err := s.b.Peek(n)
	ThrowError(err)
	return d
}

func (s *Stream) SwitchSSL() {
	config := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS13,
	}
	tlsConn := tls.Client(s.c, config)
	ThrowError(tlsConn.Handshake())
	s.c = tlsConn
	glog.Debug("switch to SSL ok")
}

func (s *Stream) PubKey() []byte {
	if c, ok := s.c.(*tls.Conn); ok {
		pub := c.ConnectionState().PeerCertificates[0].PublicKey.(*rsa.PublicKey)
		data, err := asn1ber.Marshal(*pub)
		ThrowError(err)
		return data
	}
	Throw(fmt.Errorf("not tls connection"))
	return nil
}

// ChannelBindingToken computes the channel binding token from the TLS certificate
// This implements RFC 5929 Channel Binding for TLS
func (s *Stream) ChannelBindingToken() []byte {
	if c, ok := s.c.(*tls.Conn); ok {
		// Create the channel binding token as specified in RFC 5929
		// For TLS, we use the "tls-server-end-point" channel binding type
		// The token is: SHA256(certificate)
		cert := c.ConnectionState().PeerCertificates[0]

		// Compute SHA256 hash of the certificate
		hash := sha256.Sum256(cert.Raw)
		return hash[:]
	}
	Throw(fmt.Errorf("not tls connection"))
	return nil
}

func (s *Stream) Close() {
	_ = s.c.Close()
}

// GetCurrentTimestamp returns the current timestamp in milliseconds since epoch
func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

func NewStream(addr string, tmOut time.Duration) *Stream {
	conn, err := net.DialTimeout("tcp", addr, tmOut)
	ThrowError(err)
	s := &Stream{c: conn}
	s.r = func(b []byte) (int, error) { return s.c.Read(b) }
	s.w = func(b []byte) (int, error) { return s.c.Write(b) }
	return s
}
