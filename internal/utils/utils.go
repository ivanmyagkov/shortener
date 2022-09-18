//	Package utils of auxiliary functions.
package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

//	MD5 Generating a shortened link.
func MD5(url []byte) string {
	h := md5.Sum(url)
	return fmt.Sprintf("%x", h[:8])
}

func NewURL(host string, url string) string {
	return host + "/" + url
}

//	CreateID Creating user ID.
func CreateID(size int) string {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
	}

	id := hex.EncodeToString(b)
	return id
}

func CheckIP(ip string, trustedNet string) error {
	if ip == "" {
		return interfaces.ErrNetNotTrusted
	}
	ipRequest, _, err := net.ParseCIDR(ip)
	if err != nil {
		return err
	}
	var ipnet *net.IPNet = nil
	if trustedNet != "" {
		_, ipnet, err = net.ParseCIDR(trustedNet)
		if err != nil {
			return err
		}
	}
	if ipnet.Contains(ipRequest) {
		return nil
	}
	return interfaces.ErrNetNotTrusted
}
