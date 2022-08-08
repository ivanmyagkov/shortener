//	Package of auxiliary functions.
package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

//	Generating a shortened link.
func MD5(url []byte) string {
	h := md5.Sum(url)
	return fmt.Sprintf("%x", h[:8])
}

func NewURL(host string, url string) string {
	return host + "/" + url
}

//	Creating user ID.
func CreateID(size int) string {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
	}

	id := hex.EncodeToString(b)
	return id
}
