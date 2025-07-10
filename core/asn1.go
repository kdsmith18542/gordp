// Package core provides essential utilities and data structures for the GoRDP library.
// This package includes:
//   - Generic data structures (Cache, Queue, Stack, Option, Result, Either)
//   - Error handling utilities with context and structured errors
//   - Network stream management
//   - Cryptographic utilities (MD4, HMAC-MD5, NTLM hash functions)
//   - ASN.1 parsing and serialization
//   - Async utilities for concurrent operations
//   - Type-safe generic collections and utilities
package core

import (
	"io"
)

// Asn1 https://www.ietf.org/rfc/rfc6025.html
type Asn1 struct {
	Tag    uint8
	Length int
	Value  []byte
	orig   []byte
}

func (s *Asn1) Serialize() []byte {
	return append(s.orig, s.Value...)
}

func (s *Asn1) Read(r io.Reader) []byte {
	var b byte
	ReadBE(r, &s.Tag) // read tag
	ReadBE(r, &b)     // read length

	s.orig = append(s.orig, s.Tag, b) // store
	if b&0x80 != 0 {                  // long length mode
		for left := b & 0x7f; left > 0; left-- {
			ReadBE(r, &b)
			s.orig = append(s.orig, b) // store
			s.Length = s.Length<<8 + int(b)
		}
	} else { // short length mode
		s.Length = int(b)
	}
	s.Value = make([]byte, s.Length)
	_, err := io.ReadFull(r, s.Value)
	ThrowError(err)
	return s.Serialize()
}
