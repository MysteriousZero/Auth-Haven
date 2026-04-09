package service

import (
	"auth-haven/internal/domain/interfaces"
	"crypto/rand"
	"encoding/base32"
	"fmt"

	"github.com/pquerna/otp/totp"
)

type totpGenerator struct {
	issuer string
}

func NewTOTPGenerator() interfaces.TOTPGenerator {
	return &totpGenerator{
		issuer: "Auth Haven",
	}
}

func (g *totpGenerator) GenerateSecret() string {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		// Fallback to less secure method if crypto/rand fails
		secret = []byte("DEFAULT_SECRET_CHANGE_ME")
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
}

func (g *totpGenerator) GenerateQRCodeURL(secret, email string) string {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      g.issuer,
		AccountName: email,
		Secret:      []byte(secret),
	})
	if err != nil {
		return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
			g.issuer, email, secret, g.issuer)
	}
	return key.URL()
}

