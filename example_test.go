package token

import (
	"log"
	"os"
	"testing"
	"time"

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
			RefreshTokenTTL:    1200,
			RefreshTokenPrefix: "refresh_token:",
			Timeout:            10,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Run()
	os.Exit(0)
}

func TestCookieSessionMode(t *testing.T) {
	payload := make(map[string]any)
	payload["field1"] = "value1"
	payload["field2"] = "value2"

	t.Log("---------- make access token --------------")
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
	var allPayload map[string]string
	allPayload, err = accessToken.GetAll()
	t.Log("payload", allPayload)

	t.Log("---------- waiting 3 second refresh access token --------------")
	time.Sleep(3 * time.Second)
	err = accessToken.Refresh()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("refresh_count: ", accessToken.RefreshCount())
	t.Log("refreshed_at: ", accessToken.RefreshedAt())
	t.Log("expires_at: ", accessToken.ExpiresAt())
	t.Log("payload:")
	getPayload, err := accessToken.GetAll()
	if err != nil {
		t.Error(err)
		return
	}
	for k := range getPayload {
		t.Log("    ", k, getPayload[k])
	}

	t.Log("---------- destroy access token --------------")
	err = accessToken.Destroy()
	if err != nil {
		t.Error(err)
		return
	}
	_, err = testInst.ParseAccessToken(accessToken.Value())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("access token destroyed")
}

func TestTokenMode(t *testing.T) {
	t.Log("---------- make refresh token --------------")
	payload := make(map[string]any)
	payload["field1"] = "value1"
	payload["field2"] = "value2"
	refreshToken, err := testInst.MakeRefreshToken(payload)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("refresh token value: ", refreshToken.Value())
	t.Log("refresh token created_at: ", refreshToken.CreatedAt())
	t.Log("refresh token expires_at: ", refreshToken.ExpiresAt())
	t.Log("refresh token used_count: ", refreshToken.UsedCount())
	t.Log("refresh token used_at: ", refreshToken.UsedAt())
	t.Log("refresh token payload: ", refreshToken.Payload())

	t.Log("---------- exchange access token --------------")
	accessToken, err2 := refreshToken.Exchange()
	if err2 != nil {
		t.Error(err2)
		return
	}
	t.Log("access token value: ", accessToken.Value())
	t.Log("access token created_at: ", accessToken.CreatedAt())
	t.Log("access token expires_at: ", accessToken.ExpiresAt())
	t.Log("access token refreshed_at: ", accessToken.RefreshedAt())
	t.Log("access token refresh_count: ", accessToken.RefreshCount())
	getPayload, err := accessToken.GetAll()
	if err != nil {
		t.Error(err)
		return
	}
	for k := range getPayload {
		t.Log("    ", k, getPayload[k])
	}
	t.Log("refresh token access token: ", refreshToken.AccessToken())

	t.Log("---------- destroy refresh token --------------")
	err = refreshToken.Destroy()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("---------- parse token --------------")
	_, err = testInst.ParseAccessToken(accessToken.Value())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("access token destroyed")
	_, err = testInst.ParseRefreshToken(refreshToken.Value())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("refresh token destroyed")
}
