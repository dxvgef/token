package token

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestNewRedisClient(t *testing.T) {
	NewRedisClient(nil)
}

func TestUseRedisClient(t *testing.T) {
	UseRedisClient(nil)
}

func TestSetAccessTokenPrefix(t *testing.T) {
	SetAccessTokenPrefix("access_token:")
}

func TestSetRefreshTokenPrefix(t *testing.T) {
	SetRefreshTokenPrefix("refresh_token:")
}

func TestMake(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr: "127.0.0.1:6379",
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
		Addr: "127.0.0.1:6379",
	})
	at, logicError, err := ParseAccessToken("01J5C2WEAA5T0EN5DJNMJ5Q6CM")
	if logicError != nil {
		t.Error(logicError)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.Value(), at.Payload())
}

func TestAccessToken_ExpiresAt(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	at, logicErr, err := ParseAccessToken("01J5C5887MHASFDJYP9PQJYCBY")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.ExpiresAt())
}

func TestParseRefreshToken(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	rt, logicErr, err := ParseRefreshToken("01J5C2WEADR1PWY0G5B57G4CP2")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(rt.Value(), rt.AccessToken())
}

func TestRefreshToken_ExpiresAt(t *testing.T) {
	NewRedisClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	rt, logicErr, err := ParseRefreshToken("01J5C5887QGA5ZQX5FJPN5S5WM")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(rt.ExpiresAt())
}

func TestRefreshToken_Refresh(t *testing.T) {
	var at *AccessToken
	NewRedisClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	rt, logicErr, err := ParseRefreshToken("01J5C5887QGA5ZQX5FJPN5S5WM")
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	at, logicErr, err = rt.Exchange("01J5C5887MHASFDJYP9PQJYCBY", &Options{
		AccessTokenTTL:     10 * time.Minute,
		RefreshTokenTTL:    time.Hour,
		AccessTokenPayload: "thisNewAT",
	}, true)
	if logicErr != nil {
		t.Error(logicErr)
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(at.Value(), at.Payload())
	t.Log(rt.Value(), rt.AccessToken())
}
