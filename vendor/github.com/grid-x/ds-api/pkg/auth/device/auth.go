package device

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	//PEM identifier of an RSA public key, needed for decoding
	//key content from a string
	pubKeyBlockType = "PUBLIC KEY"
)

// Verify verifies that a given byte slice was signed by the given public key
func Verify(signature, pubkey string, content []byte) error {
	hash := sha256.New()
	_, err := hash.Write(content)
	if err != nil {
		return fmt.Errorf("verification failed: %+v", err)
	}

	decodedSig, err := base64.StdEncoding.DecodeString(string(signature))
	if err != nil {
		return fmt.Errorf("verification failed: %+v", err)
	}

	block, _ := pem.Decode([]byte(pubkey))
	if block == nil || block.Type != pubKeyBlockType {
		return fmt.Errorf("verification failed")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("verification failed: %+v", err)
	}

	keyStruct, ok := key.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("verification failed")
	}

	err = rsa.VerifyPKCS1v15(keyStruct, crypto.SHA256, hash.Sum(nil), decodedSig)
	if err != nil {
		return fmt.Errorf("verification failed: %+v", err)
	}

	return nil
}

// JWTGenerator represents a generator for JWT tokens
type JWTGenerator struct {
	issuer     string
	privateKey *rsa.PrivateKey
}

// NewJWTGenerator creates a new JWTGenerator with the given settings
func NewJWTGenerator(issuer string, privateKey *rsa.PrivateKey) *JWTGenerator {
	return &JWTGenerator{
		issuer:     issuer,
		privateKey: privateKey,
	}
}

// genToken methods that takes all claims as parameters for easier testing
func (g *JWTGenerator) genToken(issuedAt, notBefore, exp time.Time, sub, accountID string) (string, error) {
	claims := struct {
		AccountID string `json:"accountID"`
		jwt.StandardClaims
	}{
		AccountID: accountID,
		StandardClaims: jwt.StandardClaims{
			Issuer:    g.issuer,
			IssuedAt:  issuedAt.Unix(),
			NotBefore: notBefore.Unix(),
			ExpiresAt: exp.Unix(),
			Subject:   sub,
		},
	}
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Sign and get the complete encoded token as a string
	return token.SignedString(g.privateKey)
}

// GenerateToken generates a token with the settings of the generator for the given subject
func (g *JWTGenerator) GenerateToken(exp time.Time, sub, accountID string) (string, error) {
	return g.genToken(time.Now(), time.Now(), exp, sub, accountID)
}

// Valid returns the claims of the token if the token is valid and an error otherwise
func (g *JWTGenerator) Valid(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return &g.privateKey.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
