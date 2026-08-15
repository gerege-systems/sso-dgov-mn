package secret

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

func box(t *testing.T) *Box {
	t.Helper()
	t.Setenv("SSO_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	b, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return b
}

func TestASealedCredentialComesBack(t *testing.T) {
	b := box(t)
	sealed, err := b.Seal("s3cr3t")
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	if bytes.Contains(sealed, []byte("s3cr3t")) {
		t.Fatal("the plaintext is still in the ciphertext")
	}
	revealed, err := b.Reveal(sealed)
	if err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if revealed != "s3cr3t" {
		t.Fatalf("got %q", revealed)
	}
}

// Two seals of the same secret must differ, or a database dump tells an
// observer which targets share a token without decrypting anything.
func TestSealingTwiceProducesDifferentCiphertext(t *testing.T) {
	b := box(t)
	first, _ := b.Seal("same")
	second, _ := b.Seal("same")
	if bytes.Equal(first, second) {
		t.Fatal("the nonce is not being used")
	}
}

func TestTamperingIsRefused(t *testing.T) {
	b := box(t)
	sealed, _ := b.Seal("s3cr3t")
	sealed[len(sealed)-1] ^= 0xff
	if _, err := b.Reveal(sealed); err == nil {
		t.Fatal("a modified ciphertext was accepted")
	}
}

// A nil Box is the state of an installation with no key, and every method has
// to answer rather than panic — the constructor of a module carries one.
func TestWithoutAKeyEveryOperationRefuses(t *testing.T) {
	t.Setenv("SSO_SECRET_KEY", "")
	if _, err := Open(); !errors.Is(err, ErrNoKey) {
		t.Fatalf("open with no key: %v", err)
	}
	var none *Box
	if _, err := none.Seal("x"); !errors.Is(err, ErrNoKey) {
		t.Fatalf("seal: %v", err)
	}
	if _, err := none.Reveal([]byte("x")); !errors.Is(err, ErrNoKey) {
		t.Fatalf("reveal: %v", err)
	}
}

func TestAKeyOfTheWrongLengthIsRefused(t *testing.T) {
	t.Setenv("SSO_SECRET_KEY", base64.StdEncoding.EncodeToString([]byte("short")))
	if _, err := Open(); err == nil {
		t.Fatal("a 5-byte key was accepted")
	}
}
