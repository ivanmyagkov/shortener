package utils

import (
	"crypto/md5"
	"fmt"
)

func MD5(url []byte) string {
	h := md5.Sum(url)
	return fmt.Sprintf("%x", h[:8])
}
