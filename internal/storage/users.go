package storage

import (
	"crypto/aes"
	"encoding/hex"
	"log"

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
	//key := utils.CreateID(16)
	return &DBUsers{
		storageUsers: map[string]ModelUser{},
		randNum:      "",
		CookieWord:   "cookie",
	}
}

func (MU *DBUsers) CreateSissionID(uid string) (string, error) {
	//generate SessionID
	MU.randNum = uid
	src, err := hex.DecodeString(MU.randNum)
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
	log.Println(cookie)
	return cookie, nil

}
func (MU *DBUsers) ReadSessionID(id string) (string, error) {
	//log.Println("read cookie = ", id)
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
	//log.Println("cid=", hex.EncodeToString(src))
	return hex.EncodeToString(src), nil
}
