package token

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

//token对象
type Token struct {
	str  string      //token字符串
	key  string      //密钥
	data interface{} //自定义数据
	exp  int64       //到期时间（unix时间戳）
	sign string      //签名
}

//token要求的参数
type Claims struct {
	Data interface{} //自定义数据
	Exp  int64       //到期时间（unix时间戳）
}

//创建一个token对象
func New(claims *Claims, key string) (*Token, error) {
	//生成一个buffer
	var claimsBuffer bytes.Buffer
	//生成序列化编码器
	encoder := gob.NewEncoder(&claimsBuffer)
	//编码claims
	err := encoder.Encode(&claims)
	if err != nil {
		return nil, err
	}
	//使用claims json + key 加密得到签名
	hash := sha256.New()
	hash.Write(claimsBuffer.Bytes())
	md := hash.Sum([]byte(key))
	sign := hex.EncodeToString(md)

	//使用base64编码claims的json格式
	claimsBase64 := base64.URLEncoding.EncodeToString(claimsBuffer.Bytes())

	//创建一个token实例
	var t Token
	t.str = claimsBase64 + "." + sign
	t.data = claims.Data
	t.exp = claims.Exp
	t.key = key
	t.sign = sign

	return &t, nil
}

//生成一个token字符串
func NewString(claims *Claims, key string) (string, error) {
	//生成一个buffer
	var claimsBuffer bytes.Buffer
	//生成序列化编码器
	encoder := gob.NewEncoder(&claimsBuffer)
	//编码claims
	err := encoder.Encode(&claims)
	if err != nil {
		return "", err
	}

	//使用claims json + key 加密得到签名
	hash := sha256.New()
	hash.Write(claimsBuffer.Bytes())
	md := hash.Sum([]byte(key))
	sign := hex.EncodeToString(md)

	//使用base64编码claims的json格式
	claimsBase64 := base64.URLEncoding.EncodeToString(claimsBuffer.Bytes())

	return claimsBase64 + "." + sign, nil
}

// 获得token对象（除str之外的属性值全部为空）
// 第二个出参可以用它的Decode()方法在外部得到claims的原始类型
func Parse(tokenStr string, key string) (*Token, error) {
	arr := strings.Split(tokenStr, ".")
	if len(arr) != 2 {
		return nil, errors.New("token格式无效")
	}
	claimsBase64 := arr[0]
	signOrig := arr[1]

	//将claims字段使用base64解码成[]byte
	claimsBytes, err := base64.URLEncoding.DecodeString(claimsBase64)
	if err != nil {
		return nil, errors.New("token格式无效")
	}

	//计算签名
	hash := sha256.New()
	hash.Write(claimsBytes)
	md := hash.Sum([]byte(key))
	sign := hex.EncodeToString(md)
	if signOrig != sign {
		return nil, errors.New("签名无效")
	}

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
		key:  key,
		sign: sign,
	}, nil
}

func (this *Token) ValidExp() bool {
	if this.exp != 0 {
		now := time.Now().Unix()
		if this.exp < now {
			return false
		}
	}
	return true
}

func (this *Token) GetString() string {
	return this.str
}

func (this *Token) GetKey() string {
	return this.key
}

func (this *Token) GetSign() string {
	return this.sign
}

func (this *Token) GetExp() int64 {
	return this.exp
}

func (this *Token) GetCustomParams() interface{} {
	return this.data
}

//快速校验token有效性，不创建token对象
func FastValid(tokenStr string, key string, checkExp bool) error {
	arr := strings.Split(tokenStr, ".")
	if len(arr) != 2 {
		return errors.New("token格式无效")
	}
	claimsBase64 := arr[0]
	signOrig := arr[1]

	//将claims字段使用base64解码成[]byte
	claimsBytes, err := base64.URLEncoding.DecodeString(claimsBase64)
	if err != nil {
		return errors.New("token格式无效")
	}

	//计算签名
	hash := sha256.New()
	hash.Write(claimsBytes)
	md := hash.Sum([]byte(key))
	sign := hex.EncodeToString(md)
	if signOrig != sign {
		return errors.New("签名无效")
	}

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
