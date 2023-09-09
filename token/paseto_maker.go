package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto/v2"
	"golang.org/x/crypto/chacha20poly1305"
)

// PasetoMaker is a pasto token maker
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker - creates a new PasetoMaker
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) < chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	return &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil

}

// CreateToken - creates token for specific username and duration
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// crating a new payload token with the given user and duration
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	// returning the signed token
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

// VerifyToken - verify the token validation
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	// decrypting the payload token into the payload var
	payload := &Payload{}
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// validating the payload
	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil

}
