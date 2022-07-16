//** Utility function to generate RSA PrivateKey and PublicKey in PEM format.
//** This can be used for testing purpose.

package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// constants for error code
const (
	ErrKeyGen = "key generation error"
	ErrPubKey = "public Key marshalling error"
)

// GenRSAKeyPair generates and returns RSA private key, public pem with error
func GenRSAKeyPair() (*rsa.PrivateKey, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, []byte{}, errors.New(ErrKeyGen)
	}

	bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, []byte{}, errors.New(ErrPubKey)
	}

	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes,
	})
	return privateKey, pem, nil
}
