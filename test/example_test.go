/*
package test

import (
	"encoding/gob"
	"testing"
	"time"

	"git.oschina.net/dxvgef/token"
)

type testClaims struct {
	UserID int64
}

var (
	key          string       = "secret" //签名密钥
	claims       testClaims              //用户数据
	claims       token.Claims            //token属性
	testTokenStr string                  //测试用的token字符串
)

func init() {
	//注册用户数据结构体
	gob.Register(testClaims{})

	//用户数据赋值
	cl.UserID = 123

	//token属性赋值
	claims.Exp = time.Now().Add(12 * time.Hour).Unix() //传入生命周期
	claims.Data = cl                                   //传入用户数据
}

//测试生成token字符串
func TestNewString(t *testing.T) {
	tokenStr, err := token.NewString(&claims, key)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(tokenStr)
}

//测试生成token对象
func TestNew(t *testing.T) {
	token, err := token.New(&claims, key)
	if err != nil {
		t.Error(err.Error())
		return
	}

	testTokenStr = token.GetString()

	t.Log(token.GetString())
}

//测试解析token字符串成token对象
func TestParse(t *testing.T) {
	//解析token字符串，得到token对象
	token, err := token.Parse(testTokenStr, key)
	if err != nil {
		t.Error(err.Error())
		return
	}

	//校验token时效
	if token.ValidExp() == false {
		t.Log("token已超时")
		return
	}
	//获得用户数据
	TokenData := token.GetData().(testClaims)

	t.Log(TokenData.UserID)
}

func TestFastValid(t *testing.T) {
	//解析token字符串，得到token对象
	err := token.FastValid(testTokenStr, key, true)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log("token字符串有效")
	}
}
*/
