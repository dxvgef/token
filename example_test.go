package token

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var testInst *Manager

func TestMain(m *testing.M) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	var err error
	testInst, err = NewManager(
		redis.NewClient(&redis.Options{
			Username: "default",
			Password: "123456",
		}),
		&ManagerOptions{
			KeyPrefix: "sess:",
			Timeout:   10,
			MakeTokenFunc: func() string {
				return strconv.FormatInt(time.Now().UnixNano(), 10)
			},
			CheckTokenFunc: func(s string) bool {
				return true
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Run()
	os.Exit(0)
}

func TestMakeAndGet(t *testing.T) {
	payload := make(map[string]any)
	payload["field1"] = "value1"
	payload["field2"] = "value2"

	t.Log("---------- make token --------------")
	testToken, err := testInst.MakeToken(&MetaData{
		TTL:          60,
		RefreshLimit: 5,
		IP:           "127.0.0.1",
		Fingerprint:  "test",
	}, payload)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("---------- meta  --------------")
	t.Log("value: ", testToken.Value())
	t.Log("ttl: ", testToken.TTL())
	t.Log("created_at: ", testToken.CreatedAt())
	t.Log("expires_at: ", testToken.ExpiresAt())
	t.Log("refreshed_at: ", testToken.RefreshedAt())
	t.Log("refresh_limit: ", testToken.RefreshLimit())
	t.Log("refreshed_count: ", testToken.RefreshedCount())
	t.Log("ip: ", testToken.IP())
	t.Log("fingerprint: ", testToken.Fingerprint())
	t.Log("child_token: ", testToken.ChildToken())
	time.Sleep(time.Second * 3)
	t.Log("---------- GetAll  --------------")
	var allPayload map[string]string
	allPayload, err = testToken.GetAll(false)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
	t.Log("---------- Get  --------------")
	var v1, v2 string
	v1, err = testToken.Get("field1")
	if err != nil {
		t.Error(err)
		return
	}
	v2, err = testToken.Get("field2")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("field1: ", v1)
	t.Log("field2: ", v2)
}

func TestParse(t *testing.T) {
	t.Log("---------- make token --------------")
	testToken, err := testInst.MakeToken(nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	t.Log("---------- parse token --------------")
	var newToken *Token
	newToken, err = testInst.ParseToken(testToken.value)
	if err != nil {
		t.Error(err)
		return
	}
	var allPayload map[string]string
	t.Log(newToken.value)
	allPayload, err = newToken.GetAll(true)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
	t.Log("---------- destroy token --------------")
	if err = testToken.Destroy(false); err != nil {
		t.Error(err)
		return
	}
}

func TestMakeAndRefresh(t *testing.T) {
	payload := make(map[string]any)
	payload["field1"] = "value1"
	payload["field2"] = "value2"

	t.Log("---------- make token --------------")
	testToken, err := testInst.MakeToken(&MetaData{
		TTL:          60,
		RefreshLimit: -1,
		IP:           "127.0.0.1",
		Fingerprint:  "test",
	}, payload)
	if err != nil {
		t.Error(err)
		return
	}
	var allPayload map[string]string
	allPayload, err = testToken.GetAll(true)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
	for i := 0; i < 4; i++ {
		time.Sleep(time.Second * 5)
		t.Log("---------- refresh  --------------")
		if err = testToken.Refresh(); err != nil {
			t.Error(err)
			return
		}
		allPayload, err = testToken.GetAll(true)
		for k := range allPayload {
			t.Log("    ", k, allPayload[k])
		}
	}
}

func TestMakeChild(t *testing.T) {
	payload := make(map[string]any)
	payload["field1"] = "value1"
	payload["field2"] = "value2"

	t.Log("---------- make token --------------")
	testToken, err := testInst.MakeToken(&MetaData{
		TTL:          60,
		RefreshLimit: 3,
		IP:           "127.0.0.1",
		Fingerprint:  "test",
	}, payload)
	if err != nil {
		t.Error(err)
		return
	}
	var allPayload map[string]string
	allPayload, err = testToken.GetAll(true)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
	time.Sleep(5 * time.Second)
	t.Log("---------- make chiled token --------------")
	newToken, err := testToken.MakeChildToken(&MetaData{
		TTL:          30,
		RefreshLimit: -1,
		IP:           "127.0.0.1",
		Fingerprint:  "test",
	}, payload)
	if err != nil {
		t.Error(err)
		return
	}
	allPayload, err = newToken.GetAll(true)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
}

func TestDestroy(t *testing.T) {
	t.Log("---------- make token --------------")
	testToken, err := testInst.MakeToken(nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	t.Log("---------- make child token --------------")
	var newToken *Token
	newToken, err = testToken.MakeChildToken(nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	var allPayload map[string]string
	t.Log(testToken.value)
	allPayload, err = testToken.GetAll(true)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
	t.Log(newToken.value)
	allPayload, err = newToken.GetAll(true)
	for k := range allPayload {
		t.Log("    ", k, allPayload[k])
	}
	time.Sleep(3 * time.Second)
	t.Log("---------- destroy parent token --------------")
	if err = testToken.Destroy(false); err != nil {
		t.Error(err)
		return
	}
}
