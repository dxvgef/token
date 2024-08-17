package token

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestMake(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "default",
		Password: "123456",
	})
	at, rt, err := Make(&Options{
		AccessTokenTTL:     10 * time.Minute,
		RefreshTokenTTL:    time.Hour,
		AccessTokenPayload: "userID",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.Value())
	t.Log(rt.Value())
}

func TestParseAccessToken(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "default",
		Password: "123456",
	})
	at, err := ParseAccessToken("01J5C2WEAA5T0EN5DJNMJ5Q6CM")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.Value(), at.Payload())
}

func TestAccessToken_ExpiresAt(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "default",
		Password: "123456",
	})
	at, err := ParseAccessToken("01J5C5887MHASFDJYP9PQJYCBY")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.ExpiresAt())
}

func TestParseRefreshToken(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "default",
		Password: "123456",
	})
	rt, err := ParseRefreshToken("01J5C2WEADR1PWY0G5B57G4CP2")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(rt.Value(), rt.AccessToken())
}

func TestRefreshToken_ExpiresAt(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "default",
		Password: "123456",
	})
	rt, err := ParseRefreshToken("01J5C5887QGA5ZQX5FJPN5S5WM")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(rt.ExpiresAt())
}

func TestRefreshToken_Refresh(t *testing.T) {
	var at *AccessToken
	NewRedisClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "default",
		Password: "123456",
	})
	rt, err := ParseRefreshToken("01J5C5887QGA5ZQX5FJPN5S5WM")
	if err != nil {
		t.Error(err)
		return
	}
	at, err = rt.Exchange("01J5C5887MHASFDJYP9PQJYCBY", &Options{
		AccessTokenTTL:     10 * time.Minute,
		RefreshTokenTTL:    time.Hour,
		AccessTokenPayload: "thisNewAT",
	}, true)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.Value(), at.Payload())
	t.Log(rt.Value(), rt.AccessToken())
}
