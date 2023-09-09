package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

// JWTMaker is a Json Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJwtMaker - creates a new JWTMaker
func NewJwtMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

// CreateToken - creates token for specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// creating the jwt token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// return the signed(with the secret key) jwt token
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken - verify the token validation
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// keyFunc - anonymous function for key validation
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// trying to convert token method into jwt.SigningMethodHMAC
		// it's should to success because we singed the token with HS256 algorithm
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	// parsing the token
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	// checking for errors - return the err message based on the error type
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// converting the jwtToken claims to Payload type
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
