/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

// Package secret encrypts the credentials this product holds on behalf of
// somebody else: a federation client secret, a SCIM bearer token.
//
// It is here rather than taken from the core because the core's equivalent
// lives in internal/, which is closed to this repository by the language. Ten
// lines of AES-GCM is the right answer to that; a fork of the platform to
// reach one function is not.
//
// The key comes from SSO_SECRET_KEY, base64, 32 bytes. Without it the product
// still starts — an installation that federates with nobody and provisions into
// nothing needs no key — but every write that would store a credential is
// refused, with the missing variable named. Storing them in the clear instead
// would be the one failure mode nobody notices until the database leaks.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

// ErrNoKey is what every operation returns when SSO_SECRET_KEY is unset.
var ErrNoKey = errors.New("SSO_SECRET_KEY is not set: this installation cannot store credentials")

// Box seals and opens the credentials this product stores.
type Box struct{ aead cipher.AEAD }

// Open reads the key from the environment.
//
// A nil Box is a working value: its methods return ErrNoKey rather than
// panicking, so the constructor of a module does not have to decide whether the
// whole product should refuse to start over a feature this installation may
// never use.
func Open() (*Box, error) {
	encoded := os.Getenv("SSO_SECRET_KEY")
	if encoded == "" {
		return nil, ErrNoKey
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("SSO_SECRET_KEY is not valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("SSO_SECRET_KEY decodes to %d bytes, want 32", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("SSO_SECRET_KEY was rejected: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("SSO_SECRET_KEY was rejected: %w", err)
	}
	return &Box{aead: aead}, nil
}

// Seal encrypts plaintext, returning nonce||ciphertext.
func (b *Box) Seal(plaintext string) ([]byte, error) {
	if b == nil {
		return nil, ErrNoKey
	}
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("could not read a nonce: %w", err)
	}
	return b.aead.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

// Reveal decrypts what Seal produced.
//
// Called on the path that uses a credential — signing in through a federated
// provider, pushing a user into a SCIM target — and never on the path that
// answers an API request. Nothing in this product returns a stored credential
// to a caller.
func (b *Box) Reveal(sealed []byte) (string, error) {
	if b == nil {
		return "", ErrNoKey
	}
	size := b.aead.NonceSize()
	if len(sealed) < size {
		return "", errors.New("the stored credential is too short to be one")
	}
	plaintext, err := b.aead.Open(nil, sealed[:size], sealed[size:], nil)
	if err != nil {
		return "", fmt.Errorf("the stored credential did not decrypt: %w", err)
	}
	return string(plaintext), nil
}
