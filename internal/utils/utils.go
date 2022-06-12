package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
)

func MD5(url []byte) string {
	h := md5.Sum(url)
	return fmt.Sprintf("%x", h[:8])
}

func NewURL(host string, url string) string {
	return host + "/" + url
}

func HashUser(userName string) []byte {
	hash := sha256.New()
	hash.Write([]byte(userName))
	return hash.Sum(nil)
}

func CreateID(size int) string {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
	}

	id := hex.EncodeToString(b)
	return id
}
