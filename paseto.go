package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"kmfg.dev/tiny/auth"

	"google.golang.org/protobuf/proto"
)

const (
	PASETO_VERSION_ENV    = "v4.local."
	PASETO_VALID_DURATION = 24 * time.Hour
)

var (
	PASETO_ENC_KEY           = GenerateSecret()
	X_CHA_CHA, X_CHA_CHA_ERR = chacha20poly1305.NewX(PASETO_ENC_KEY)
)

func ValidateXChaCha() {
	if X_CHA_CHA_ERR != nil {
		WEB_LOGGER.Fatal().Err(X_CHA_CHA_ERR).Msg("Cannot initialize xchacha20poly1305.")
	}
}

func ToPaseto(user *User) (string, error) {
	payload := &auth.UserPayload{
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Unix(),
		ExpiresAt: time.Now().Add(PASETO_VALID_DURATION).Unix(),
	}

	data, err := proto.Marshal(payload)
	if err != nil {
		return "", err
	}

	encrypted, err := Encrypt(data)
	if err != nil {
		return "", err
	}

	return PASETO_VERSION_ENV + encrypted, nil
}

func FromPaseto(paseto string) (*auth.UserPayload, error) {
	if !strings.HasPrefix(paseto, PASETO_VERSION_ENV) {
		return nil, fmt.Errorf("%s does not match the expected version and environment.", paseto)
	}

	payload := paseto[len(PASETO_VERSION_ENV):]
	decryptedPayload, err := Decrypt(payload)
	if err != nil {
		return nil, err
	}

	var userPayload auth.UserPayload
	err = proto.Unmarshal(decryptedPayload, &userPayload)

	if err != nil {
		return nil, err
	}

	if userPayload.ExpiresAt <= time.Now().Unix() {
		return nil, fmt.Errorf("PASETO is valid, but expired at %d", userPayload.ExpiresAt)
	}

	return &userPayload, nil
}

func Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, X_CHA_CHA.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := X_CHA_CHA.Seal(nil, nonce, plaintext, nil)

	combined := append(nonce, ciphertext...)
	return base64.URLEncoding.EncodeToString(combined), nil
}

func Decrypt(encrypted string) ([]byte, error) {
	combined, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	nonceSize := X_CHA_CHA.NonceSize()
	if len(combined) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	return X_CHA_CHA.Open(nil, nonce, ciphertext, nil)
}
