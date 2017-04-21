package token

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
)

//编码
func encode(claims *Claims, key string) (tokenStr string, signStr string, err error) {
	//生成一个buffer
	var claimsBuffer bytes.Buffer
	//生成序列化编码器
	encoder := gob.NewEncoder(&claimsBuffer)
	//编码claims
	err = encoder.Encode(&claims)
	if err != nil {
		return
	}

	//使用claims json + key 加密得到签名
	hash := sha256.New()
	hash.Write(claimsBuffer.Bytes())
	md := hash.Sum([]byte(key))
	signStr = hex.EncodeToString(md)

	//使用base64编码claims的json格式
	claimsBase64 := base64.URLEncoding.EncodeToString(claimsBuffer.Bytes())
	tokenStr = claimsBase64 + "." + signStr
	return
}
