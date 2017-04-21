package token

import (
	"encoding/base64"
	"encoding/json"
)

//传入claims和key，返回一个实例化的token指针
func New(claims Claims, key string) (*Token, error) {
	var token Token
	sign, tokenStr, err := operation(&claims, key)
	if err != nil {
		return nil, err
	}
	token.str = tokenStr
	token.Claims = claims
	token.sign = sign
	return &token, nil
}

//传入claims和key，返回token字符串
func NewString(claims Claims, key string) (tokenStr string, err error) {
	_, tokenStr, err = operation(&claims, key)
	if err != nil {
		return
	}
	return
}

//对claims进行编码，并计算出签名和token字符串
func operation(claims *Claims, key string) (sign string, tokenStr string, err error) {
	//将claims序列化成json
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return
	}
	//计算claims的签名
	sign, err = sha256Encode(claimsBytes, key)
	if err != nil {
		return
	}

	//用base64将claims进行编码
	claimsBase64 := base64.URLEncoding.EncodeToString(claimsBytes)
	//去掉base64结尾的==符号再拼接sign成为token字符串
	tokenStr = claimsBase64[:len(claimsBase64)-2] + "." + sign
	return
}
