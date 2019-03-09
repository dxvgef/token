package token

import "time"

// claims的方法
type Claims interface {
	Activated() bool // 检查是否激活
	Expired() bool   // 检查是否到期
	Valid() bool     // 检查所有时间决定token是否有效
	GetAT() int64    // 获得生效时间
	GetExp() int64   // 获得失效时间
}

// claims的属性
type ClaimsAttr struct {
	ClaimsAT  int64 `json:"claims_at,omitempty"`  // 生效时间
	ClaimsExp int64 `json:"claims_exp,omitempty"` // 到期时间
}

// 检查是否生效
func (this ClaimsAttr) Activated() bool {
	if this.ClaimsAT < time.Now().Unix() {
		return true
	}
	return false
}

// 检查是否到期
func (this ClaimsAttr) Expired() bool {
	if this.ClaimsExp == 0 || this.ClaimsExp > time.Now().Unix() {
		return false
	}
	return true
}

// 检查所有时间决定token是否有效
func (this ClaimsAttr) Valid() bool {
	now := time.Now().Unix()
	if this.ClaimsAT < now && this.ClaimsExp > now || this.ClaimsExp == 0 {
		return true
	}
	return false
}

// 获得生效时间
func (this ClaimsAttr) GetAT() int64 {
	return this.ClaimsAT
}

// 获得失效时间
func (this ClaimsAttr) GetExp() int64 {
	return this.ClaimsExp
}
