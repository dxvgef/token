package token

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256算法计算签名
func sha256Encode(claimsBytes []byte, key string) (sign string, err error) {
	hash := sha256.New()
	_, err = hash.Write(claimsBytes)
	if err != nil {
		return
	}
	//加密时带上key
	hashBytes := hash.Sum([]byte(key))
	//计算出字符串格式的签名
	sign = hex.EncodeToString(hashBytes)
	claimsBytes = nil
	return
}
