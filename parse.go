package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
)

// 解析token字符串并返回token
func Parse(tokenStr string, claims Claims, key string) (*Token, error) {
	var token Token
	var err error

	// 将tokenStr按.符号拆成字符串数组
	arr := strings.Split(tokenStr, ".")
	// 判断数组长度及值是否为空
	if len(arr) != 2 {
		return nil, errors.New("token格式不正确")
	}
	if arr[0] == "" || arr[1] == "" {
		return nil, errors.New("token格式不正确")
	}

	// 将数组的claims部份用base64解码
	claimsBytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(arr[0])
	if err != nil {
		return nil, err
	}

	// 计算claims的签名
	sign, err := sha256Encode(claimsBytes, key)
	if err != nil {
		return nil, err
	}
	// 对比签名
	if sign != arr[1] {
		return nil, errors.New("签名校验失败")
	}

	// 将claims使用json反序列化到函数传入的claims中
	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	token.Claims = claims
	token.sign = sign
	token.str = tokenStr
	return &token, nil
}
