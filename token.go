package token

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"
)

//token对象
type Token struct {
	str  string      //token字符串
	data interface{} //自定义数据
	exp  int64       //到期时间（unix时间戳）
	sign string      //签名
}

//token要求的参数
type Claims struct {
	Data interface{} //自定义数据
	Exp  int64       //到期时间（unix时间戳）
}

//创建一个token对象，可使用String()方法获得token字符串
func New(claims *Claims, key string) (*Token, error) {
	tokenStr, signStr, err := encode(claims, key)
	var t Token
	t.str = tokenStr
	t.data = claims.Data
	t.exp = claims.Exp
	t.sign = signStr
	return &t, err
}

//生成一个token字符串，不返回token对象
func NewString(claims *Claims, key string) (string, error) {
	tokenStr, _, err := encode(claims, key)
	return tokenStr, err
}

// 解析token字符串并返回token对象
func Parse(tokenStr string, key string) (*Token, error) {
	claimsBytes, sign, err := decode(tokenStr, key)
	//反序列化claims
	var claims Claims
	buffer := bytes.NewBuffer(claimsBytes)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&claims)
	if err != nil {
		return nil, err
	}

	return &Token{
		str:  tokenStr,
		exp:  claims.Exp,
		data: claims.Data,
		sign: sign,
	}, nil
}

//快速解析token字符串有效性，不返回token对象
//checkExp决定是否校验时间，时间为0认为无效
func FastValid(tokenStr string, key string, checkExp bool) error {
	claimsBytes, _, err := decode(tokenStr, key)
	//检查exp
	if checkExp == true {
		//反序列化claims
		var claims Claims
		buffer := bytes.NewBuffer(claimsBytes)
		decoder := gob.NewDecoder(buffer)
		err = decoder.Decode(&claims)
		if err != nil {
			return err
		}

		if claims.Exp == 0 {
			return errors.New("token没有失效时间")
		} else {
			now := time.Now().Unix()
			if claims.Exp < now {
				return errors.New("token已失效")
			}
		}
	}

	return nil
}

//校验token时效
func (this *Token) ValidExp() bool {
	if this.exp != 0 {
		now := time.Now().Unix()
		if this.exp < now {
			return false
		}
	}
	return true
}

//获得token字符串
func (this *Token) GetString() string {
	return this.str
}

//获得签名
func (this *Token) GetSign() string {
	return this.sign
}

//获得到期时间
func (this *Token) GetExp() int64 {
	return this.exp
}

//获得数据
func (this *Token) GetData() interface{} {
	return this.data
}
