package token

import (
	"log"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

var testInst *Token

func TestMain(m *testing.M) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	var err error
	testInst, err = New(
		redis.NewClient(&redis.Options{
			Username: "default",
			Password: "123456",
		}),
		&Options{
			AccessTokenTTL:     600,
			AccessTokenPrefix:  "access_token:",
			CookieSessionMode:  false,
			RefreshTokenTTL:    1200,
			RefreshTokenPrefix: "refresh_token:",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Run()
	os.Exit(0)
}

func TestToken_MakeAccessToken(t *testing.T) {
	payload := make(map[string]string)
	payload["field1"] = "字段1"
	payload["field2"] = "字段2"
	accessToken, err := testInst.MakeAccessToken(payload)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("value: ", accessToken.Value())
	t.Log("created_at: ", accessToken.CreatedAt())
	t.Log("expires_at: ", accessToken.ExpiresAt())
	t.Log("refreshed_at: ", accessToken.RefreshedAt())
	t.Log("refresh_count: ", accessToken.RefreshCount())
	for k := range accessToken.payload {
		t.Logf("%s: %s", k, accessToken.payload[k])
	}
}

func TestToken_ParseAccessToken(t *testing.T) {
	accessToken, logicErr, runtimeErr := testInst.ParseAccessToken("01J6J1YE7GPVPFX306XDH3XPEF")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if runtimeErr != nil {
		t.Error(logicErr)
		return
	}
	t.Log("value: ", accessToken.Value())
	t.Log("created_at: ", accessToken.CreatedAt())
	t.Log("expires_at: ", accessToken.ExpiresAt())
	t.Log("refreshed_at: ", accessToken.RefreshedAt())
	t.Log("refresh_count: ", accessToken.RefreshCount())
	for k := range accessToken.payload {
		t.Logf("%s: %s", k, accessToken.payload[k])
	}
}

func TestAccessToken_Refresh(t *testing.T) {
	accessToken, logicErr, runtimeErr := testInst.ParseAccessToken("01J6J3E3RA7WQXDE4S53BATNPR")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if runtimeErr != nil {
		t.Error(logicErr)
		return
	}
	t.Log("value: ", accessToken.Value())
	t.Log("created_at: ", accessToken.CreatedAt())
	t.Log("expires_at: ", accessToken.ExpiresAt())
	t.Log("refreshed_at: ", accessToken.RefreshedAt())
	t.Log("refresh_count: ", accessToken.RefreshCount())
	for k := range accessToken.payload {
		t.Logf("%s: %s", k, accessToken.payload[k])
	}
	t.Log("---------------------------------")
	err := accessToken.Refresh()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("created_at: ", accessToken.CreatedAt())
	t.Log("expires_at: ", accessToken.ExpiresAt())
	t.Log("refreshed_at: ", accessToken.RefreshedAt())
	t.Log("refresh_count: ", accessToken.RefreshCount())
}

func TestAccessToken_Destroy(t *testing.T) {
	accessToken, logicErr, runtimeErr := testInst.ParseAccessToken("01J6J68R4VZX93P9080W1ZS64F")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if runtimeErr != nil {
		t.Error(logicErr)
		return
	}
	err := accessToken.Destroy()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestToken_MakeRefreshToken(t *testing.T) {
	refreshToken, err := testInst.MakeRefreshToken("01J6J68R4VZX93P9080W1ZS64F")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("value: ", refreshToken.Value())
	t.Log("access token: ", refreshToken.AccessToken())
	t.Log("created_at: ", refreshToken.CreatedAt())
	t.Log("expires_at: ", refreshToken.ExpiresAt())
	t.Log("use_count: ", refreshToken.UseCount())
	t.Log("used_at: ", refreshToken.UsedAt())
}

func TestToken_ParseRefreshToken(t *testing.T) {
	refreshToken, logicErr, runtimeErr := testInst.ParseRefreshToken("01J6J2KWX8XZH1XD1XQMJVRDV2")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if runtimeErr != nil {
		t.Error(runtimeErr)
		return
	}
	t.Log("value: ", refreshToken.Value())
	t.Log("access token: ", refreshToken.AccessToken())
	t.Log("created_at: ", refreshToken.CreatedAt())
	t.Log("expires_at: ", refreshToken.ExpiresAt())
	t.Log("use_count: ", refreshToken.UseCount())
	t.Log("used_at: ", refreshToken.UsedAt())
}

func TestRefreshToken_Exchange(t *testing.T) {
	refreshToken, logicErr, runtimeErr := testInst.ParseRefreshToken("01J6J572GRG8H3444RTGP8V85B")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if runtimeErr != nil {
		t.Error(logicErr)
		return
	}

	payload := make(map[string]string)
	payload["fieldA"] = "字段A"
	payload["fieldB"] = "字段B"
	accessToken, lErr, rErr := refreshToken.Exchange(payload)
	if lErr != nil {
		t.Error(lErr)
		return
	}
	if rErr != nil {
		t.Error(rErr)
		return
	}
	t.Log("value: ", accessToken.Value())
	t.Log("created_at: ", accessToken.CreatedAt())
	t.Log("expires_at: ", accessToken.ExpiresAt())
	t.Log("refreshed_at: ", accessToken.RefreshedAt())
	t.Log("refresh_count: ", accessToken.RefreshCount())
	for k := range accessToken.payload {
		t.Logf("%s: %s", k, accessToken.payload[k])
	}
	t.Log("---------------------------")
	t.Log("used_at: ", refreshToken.UsedAt())
	t.Log("use_count: ", refreshToken.UseCount())
}

func TestRefreshToken_Destroy(t *testing.T) {
	refreshToken, logicErr, runtimeErr := testInst.ParseRefreshToken("01J6J69C3BM2NNFFDP99GHEDSZ")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if runtimeErr != nil {
		t.Error(logicErr)
		return
	}

	err := refreshToken.Destroy(false)
	if err != nil {
		t.Error(err)
		return
	}
}
