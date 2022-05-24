package storage

import (
	"crypto/aes"
	"encoding/hex"

	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
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
	key := utils.CreateID(16)
	return &DBUsers{
		storageUsers: map[string]ModelUser{},
		randNum:      key,
		CookieWord:   "cookie",
	}
}

func (MU *DBUsers) CreateSissionID() (string, error) {
	//generate SessionID
	id := MU.randNum
	src, err := hex.DecodeString(id)
	if err != nil {
		return "", err
	}
	//read secret key
	key := []byte(config.Secret)

	//sign the session whis a secret key
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	dst := make([]byte, 16)

	aesblock.Encrypt(dst, src)
	cookie := hex.EncodeToString(dst)
	return cookie, nil

}
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
	//decryption session id by secret key
	src := make([]byte, 16)
	aesblock.Decrypt(src, dst)

	return hex.EncodeToString(src), nil
}
