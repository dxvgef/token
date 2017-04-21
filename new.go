package token

import (
	"encoding/base64"
	"encoding/json"
)

//传入claims和key，返回一个实例化的token指针
func New(claims Claims, key string) (*Token, error) {
	var token Token
	//将claims序列化成json
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}
	//计算claims的签名
	sign, err := sha256Encode(claimsBytes, key)
	if err != nil {
		return nil, err
	}

	//将claims的base64字符串与签名拼接出token字符串
	claimsBase64 := base64.URLEncoding.EncodeToString(claimsBytes)
	token.str = claimsBase64 + "." + sign

	token.Claims = claims
	token.sign = sign
	return &token, nil
}
