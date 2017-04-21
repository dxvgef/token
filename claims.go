package token

import "time"

//claims的方法
type Claims interface {
	Activated() bool //检查是否激活
	Expired() bool   //检查是否到期
	Valid() bool     //检查所有时间决定token是否有效
}

// claims的属性
type ClaimsAttr struct {
	ClaimsAT  int64 `json:"claims_at,omitempty"`  //激活时间
	ClaimsExp int64 `json:"claims_exp,omitempty"` //到期时间
}

//检查是否激活
func (this ClaimsAttr) Activated() bool {
	if this.ClaimsAT < time.Now().Unix() {
		return true
	}
	return false
}

//检查是否到期
func (this ClaimsAttr) Expired() bool {
	if this.ClaimsExp < time.Now().Unix() {
		return true
	}
	return false
}

//检查所有时间决定token是否有效
func (this ClaimsAttr) Valid() bool {
	now := time.Now().Unix()
	if this.ClaimsAT > now && this.ClaimsExp < now {
		return true
	}
	return false
}
