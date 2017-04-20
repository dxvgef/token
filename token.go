package token

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Token struct {
	str    string
	Claims *Claims
}
type Claims struct {
	Exp    int64
	UserID int64
}

func New(claims *Claims, key string) (*Token, error) {
	//对claims序列化
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}
	//使用claims json + key 加密得到签名
	claimsStr := string(claimsJSON)
	signBytes := []byte(claimsStr + "$" + key)
	hash := sha256.New()
	hash.Write(signBytes)
	md := hash.Sum(nil)
	sign := hex.EncodeToString(md)

	//使用base64编码claims的json格式
	claimsBase64 := base64.URLEncoding.EncodeToString(claimsJSON)

	var wt Token
	wt.Claims = claims
	wt.str = claimsBase64 + "." + sign
	return &wt, nil
}

func Parse(tokenStr string, key string) (*Token, error) {
	tokenArr := strings.Split(tokenStr, ".")
	if len(tokenArr) != 2 {
		return nil, errors.New("token格式无效")
	}
	claimsBase64 := tokenArr[0]
	signOrig := tokenArr[1]
	claimsBytes, err := base64.URLEncoding.DecodeString(claimsBase64)
	if err != nil {
		return nil, errors.New("token格式无效[base64]")
	}
	claimsStr := string(claimsBytes)

	signBytes := []byte(claimsStr + "$" + key)

	hash := sha256.New()
	hash.Write(signBytes)
	md := hash.Sum(nil)
	sign := hex.EncodeToString(md)

	if signOrig != sign {
		return nil, errors.New("签名无效")
	}

	var claims Claims
	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		return nil, errors.New("claims格式无效")
	}

	var wt Token
	wt.Claims = &claims
	wt.str = tokenStr
	return &wt, nil
}

func (this *Token) ValidExp() bool {
	if this.Claims.Exp != 0 {
		now := time.Now().Unix()
		if this.Claims.Exp < now {
			return false
		}
	}
	return true
}

func (this *Token) String() string {
	return this.str
}
