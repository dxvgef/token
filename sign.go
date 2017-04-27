package token

import (
	"crypto/sha256"
	"encoding/hex"
	"unsafe"
)

// SHA256算法计算签名
func sha256Encode(claimsBytes []byte, key string) (sign string, err error) {
	hash := sha256.New()
	_, err = hash.Write(claimsBytes)
	if err != nil {
		return
	}
	//加密时带上key
	hash.Write(strToByte(key))
	//计算出字符串格式的签名
	sign = hex.EncodeToString(hash.Sum(nil))
	claimsBytes = nil
	return
}

func strToByte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
