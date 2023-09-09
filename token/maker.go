package token

import (
	"fmt"
	"strings"
	"time"
)

const (
	JWT    = "JWT"
	PASETO = "PASETO"
)

// Maker is an interface for managing token
type Maker interface {
	// CreateToken - creates token for specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)
	// VerifyToken - verify the token validation
	VerifyToken(token string) (*Payload, error)
}

// this function generate a maker for the given token type and signed it with the secret key
func CreateNewToken(tokenType, secretKey string) (Maker, error) {
	switch strings.ToUpper(tokenType) {
	case JWT:
		return NewJwtMaker(secretKey)
	case PASETO:
		return NewPasetoMaker(secretKey)
	default:
		return nil, fmt.Errorf("%s is unknown token type", tokenType)
	}
}
