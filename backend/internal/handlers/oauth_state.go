package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"
)

func GenerateSignedState() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	payload := base64.URLEncoding.EncodeToString(nonce) + "." + ts
	mac := sign(payload)

	return payload + "." + mac, nil
}

// verifySignedState returns true if the state has a valid signature
// and was generated within the last 5 minutes.
func VerifySignedState(state string) bool {
	parts := strings.SplitN(state, ".", 3)
	if len(parts) != 3 {
		return false
	}

	payload := parts[0] + "." + parts[1]
	expectedMAC := sign(payload)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedMAC)) {
		return false
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}

	age := time.Since(time.Unix(ts, 0))
	return age < 5*time.Minute
}

func sign(payload string) string {
	secret := []byte(os.Getenv("JWT_SECRET"))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
