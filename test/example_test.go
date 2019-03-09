package test

import (
	"testing"
	"time"

	"github.com/dxvgef/token"
)

type testClaims struct {
	UserID int64
	token.ClaimsAttr
}

var (
	key          string = "secret" // 签名密钥
	claims       testClaims
	testTokenStr string // 测试用的token字符串
)

func init() {
	claims.UserID = 123
	claims.ClaimsAT = time.Now().Add(5 * time.Second).Unix()   // 激活时间
	claims.ClaimsExp = time.Now().Add(10 * time.Second).Unix() // 到期时间
}

// 测试生成token字符串
func TestNewString(t *testing.T) {
	tokenStr, err := token.NewString(claims, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(tokenStr)
}

// 测试生成token对象
func TestNew(t *testing.T) {
	tk, err := token.New(&claims, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	testTokenStr = tk.GetString()
	t.Log(testTokenStr)
}

// 测试解析token字符串成token对象
func TestParse(t *testing.T) {
	// 解析token字符串，得到token对象
	tk, err := token.Parse(testTokenStr, &testClaims{}, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// 获得claims数据
	claims := tk.Claims.(*testClaims)
	t.Log(claims.UserID)

	// 检查是否激活
	if claims.Activated() == true {
		t.Log("token已激活")
	} else {
		t.Log("token还未激活时间")
	}

	// 检查是否到期
	if claims.Expired() == true {
		t.Log("token已到期")
	} else {
		t.Log("token未到期")
	}

	// 检查是否有效，会同时检查激活时间和到期时间
	if claims.Valid() == true {
		t.Log("token有效")
	} else {
		t.Log("token无效")
	}
}
