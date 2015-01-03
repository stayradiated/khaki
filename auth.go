package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/dustin/randbo"
	"github.com/paypal/gatt"
)

type Auth struct {
	authed        bool
	lastChallenge []byte
	SecretKey     []byte
}

func NewAuth(secretKey []byte) *Auth {
	return &Auth{
		lastChallenge: make([]byte, 16),
		authed:        false,
		SecretKey:     secretKey,
	}
}

func (a *Auth) IsAuthenticated() bool {
	return a.authed
}

func (a *Auth) HandleAuthRead(resp gatt.ReadResponseWriter, req *gatt.ReadRequest) {
	randbo.New().Read(a.lastChallenge)
	resp.Write(a.lastChallenge)
	fmt.Println(a.lastChallenge)
}

func (a *Auth) HandleAuthWrite(req gatt.Request, data []byte) (status byte) {
	mac := hmac.New(sha256.New, a.SecretKey)
	mac.Write(a.lastChallenge)
	expectedMac := mac.Sum(nil)
	equal := hmac.Equal(expectedMac, data)

	a.authed = equal

	fmt.Println("data:", data)
	fmt.Println("hmac:", expectedMac)

	fmt.Println("Authentication status:", equal)

	return gatt.StatusSuccess
}
