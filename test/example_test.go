package test

import (
	"encoding/gob"
	"testing"
	"time"

	"git.oschina.net/dxvgef/token"
)

type Data struct {
	UserID int64
}

var (
	key          string       = "secret" //签名密钥
	data         Data                    //用户数据
	claims       token.Claims            //token属性
	testTokenStr string                  //测试用的token字符串
)

func init() {
	//注册用户数据结构体
	gob.Register(Data{})

	//用户数据赋值
	data.UserID = 123

	//token属性赋值
	claims.Exp = time.Now().Add(12 * time.Hour).Unix() //传入生命周期
	claims.Data = data                                 //传入用户数据
}

//测试生成token字符串
func Test_NewString(t *testing.T) {
	tokenStr, err := token.NewString(&claims, key)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(tokenStr)
}

//测试生成token对象
func Test_New(t *testing.T) {
	token, err := token.New(&claims, key)
	if err != nil {
		t.Error(err.Error())
		return
	}

	testTokenStr = token.GetString()

	t.Log(token.GetString())
}

//测试解析token字符串成token对象
func Test_Parse(t *testing.T) {
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
	TokenData := token.GetData().(Data)

	t.Log(TokenData.UserID)
}

func Test_FastValid(t *testing.T) {
	//解析token字符串，得到token对象
	err := token.FastValid(testTokenStr, key, true)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log("token字符串有效")
	}
}
