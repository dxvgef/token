**生成Token字符串**

	//自定义结构体
	type TokenData struct {
    	UserID int64
    }

    // 在gob包里注册自定义结构体
	gob.Register(TokenData{})

	//生成签名算法的密钥字符串
	var key string = "secret"
	
	//实例化一个claims
	var claims token.Claims
	//claims的数据是一个自定义结构体
	claims.Data = TokenData{
		UserID: 123,
	}
	//生命周期
	claims.Exp = time.Now().Unix()

	//生成一个token字符串
	tokenStr, err := token.NewString(&claims, key)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("token字符串:", tokenStr)


**解析Token字符串**

	//解析token字符串并获得token对象
	t, err := token.Parse(tokenStr, key)
	if err != nil {
		log.Println(err.Error())
		return
	}
	//将获得的数据断言为自定义结构体类型
	data := t.GetData().(TokenData)

	//打印token对象的各属性
	log.Println(data.UserID)
	log.Println(t.GetString())
	log.Println(t.GetExp())
	log.Println(t.GetKey())

	//检查超时
	if t.ValidExp() == false {
		log.Println("token超时")
	} else {
		log.Println("token未超时")
	}


**快速校验token字符串是否有效**

	err := token.FastValid(tokenStr, key, true)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("token有效")