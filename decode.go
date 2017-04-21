package token

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

//解码
func decode(tokenStr string, key string) (claimsBytes []byte, sign string, err error) {
	arr := strings.Split(tokenStr, ".")
	if len(arr) != 2 {
		err = errors.New("token格式无效")
		return
	}
	claimsBase64 := arr[0]
	signOrig := arr[1]

	//将claims字段使用base64解码成[]byte
	claimsBytes, err = base64.URLEncoding.DecodeString(claimsBase64)
	if err != nil {
		err = errors.New("token格式无效")
		return
	}

	//计算签名
	hash := sha256.New()
	hash.Write(claimsBytes)
	md := hash.Sum([]byte(key))
	sign = hex.EncodeToString(md)
	if signOrig != sign {
		err = errors.New("签名无效")
		return
	}
	return
}
