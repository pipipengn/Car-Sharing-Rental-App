package token

import (
	"crypto/rsa"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTTokenGen generates a JWT token.
type JWTTokenGen struct {
	privateKey *rsa.PrivateKey
	issuer     string
	nowFunc    func() time.Time
}

// NewJWTTokenGen creates a JWTTokenGen.
func NewJWTTokenGen(issuer string, privateKey *rsa.PrivateKey) *JWTTokenGen {
	return &JWTTokenGen{
		issuer:     issuer,
		nowFunc:    time.Now,
		privateKey: privateKey,
	}
}

// GenerateToken generates a token.
func (t *JWTTokenGen) GenerateToken(accountID string, expire time.Duration) (string, error) {
	nowSec := t.nowFunc().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.StandardClaims{
		Issuer:    t.issuer,
		IssuedAt:  nowSec,
		ExpiresAt: nowSec + int64(expire.Seconds()),
		Subject:   accountID,
	})

	return token.SignedString(t.privateKey)
}

func PrivateKey(keyPath string) *rsa.PrivateKey {
	pkFile, err := os.Open(keyPath)
	if err != nil {
		log.Fatal("cannot open private key", zap.Error(err))
	}

	pkBytes, err := ioutil.ReadAll(pkFile)
	if err != nil {
		log.Fatal("cannot read private key", zap.Error(err))
	}

	privteKey, err := jwt.ParseRSAPrivateKeyFromPEM(pkBytes)
	if err != nil {
		log.Fatal("cannot parse private key", zap.Error(err))
	}

	return privteKey
}
