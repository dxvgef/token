package token

type Token struct {
	Claims Claims // 要求的内容
	sign   string // 签名
	str    string // token字符串
}

// 获得签名
func (this *Token) GetSign() string {
	return this.sign
}

// 获得token字符串
func (this *Token) GetString() string {
	return this.str
}
