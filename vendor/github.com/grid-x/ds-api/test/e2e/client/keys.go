package client

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// Sign signes some request data with a given private key
func Sign(data []byte, privkey *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.New()
	if _, err := bytes.NewReader(data).WriteTo(hash); err != nil {
		return nil, err
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, privkey, crypto.SHA256, hash.Sum(nil))
	if err != nil {
		return nil, err
	}

	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(sig)))
	base64.StdEncoding.Encode(b64, sig)

	return b64, nil
}

// GetPublicKeyPem extracts a public key in PEM format from private key
func GetPublicKeyPem(privKey *rsa.PrivateKey) (string, error) {
	publicKeySerialised, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", err
	}
	var publicKey bytes.Buffer
	if err := pem.Encode(&publicKey, &pem.Block{Type: "PUBLIC KEY", Bytes: publicKeySerialised}); err != nil {
		return "", err
	}

	return publicKey.String(), nil
}

// GetPrivKeyFromPem extracts a RSA private key from a given PEM formatted input stream
func GetPrivKeyFromPem(data []byte) (*rsa.PrivateKey, error) {
	block, err := extractBlock(data, "RSA PRIVATE KEY")
	if err != nil {
		return nil, err
	}

	k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// ExtractBlock extracts the given block type from the pem encoded data
func extractBlock(data []byte, blockType string) (*pem.Block, error) {
	var block *pem.Block
	oldLength := len(data)
	for {
		block, data = pem.Decode(data)
		if block == nil {
			if len(data) == 0 {
				return nil, fmt.Errorf("failed to decode PEM block %s", blockType)
			}
			if len(data) == oldLength {
				return nil, fmt.Errorf("malformed PEM input")
			}
		} else if block.Type == blockType {
			return block, nil
		}
		oldLength = len(data)
	}
}
