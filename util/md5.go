package util

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(bytes []byte) string {
	hasher := md5.New()
	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
