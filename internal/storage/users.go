package storage

import (
	"crypto/aes"
	"encoding/hex"

	"github.com/ivanmyagkov/shortener.git/internal/config"
)

type DBUsers struct {
	storageUsers map[string]ModelUser
	randNum      string
	CookieWord   string
}

type ModelUser struct {
	userID string
	cookie string
}

func New() *DBUsers {
	return &DBUsers{
		storageUsers: map[string]ModelUser{},
		randNum:      "",
		CookieWord:   "cookie",
	}
}

//	CreateSissionID Creating a session id for cookies.
func (MU *DBUsers) CreateSissionID(uid string) (string, error) {
	// Generate SessionID
	MU.randNum = uid
	src, err := hex.DecodeString(MU.randNum)
	if err != nil {
		return "", err
	}
	// Read secret key
	key := []byte(config.Secret)

	// Sign the session with a secret key
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	dst := make([]byte, 16)

	aesblock.Encrypt(dst, src)
	cookie := hex.EncodeToString(dst)
	return cookie, nil

}

//	ReadSessionID Reading the user ID.
func (MU *DBUsers) ReadSessionID(id string) (string, error) {
	key := []byte(config.Secret)
	dst, err := hex.DecodeString(id)
	if err != nil {
		return "", err
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// Decryption session id by secret key
	src := make([]byte, 16)
	aesblock.Decrypt(src, dst)
	return hex.EncodeToString(src), nil
}
