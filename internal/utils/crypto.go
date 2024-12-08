package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

type RSA struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

var keys RSA

// GenerateRSAKeys generates an RSA key pair.
func GenerateRSAKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	keys = RSA{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	return nil
}

// Encrypt encrypts a message using the public key.
func Encrypt(message string) (string, error) {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		keys.publicKey,
		[]byte(message),
		nil,
	)
	if err != nil {
		return "", err
	}

	// Encode to base64 for easier transport
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// Decrypt decrypts a message using the private key.
func Decrypt(encryptedMessage string) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		keys.privateKey,
		encryptedBytes,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}
