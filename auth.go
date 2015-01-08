package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/dustin/randbo"
	"github.com/paypal/gatt"
)

type Auth struct {
	isAuthed  bool
	public    bool
	challenge []byte
	Secret    []byte
}

type AuthConfig struct {
	Secret []byte
	Public bool
}

func NewAuth(c *AuthConfig) *Auth {
	return &Auth{
		challenge: make([]byte, 16),
		isAuthed:  false,
		public:    c.Public,
		Secret:    c.Secret,
	}
}

func (a Auth) IsAuthenticated() bool {
	if a.public {
		return true
	} else {
		return a.isAuthed
	}
}

func (a *Auth) HandleAuthRead(resp gatt.ReadResponseWriter, req *gatt.ReadRequest) {
	randbo.New().Read(a.challenge)
	resp.Write(a.challenge)
	fmt.Printf("Creating new challenge: %x\n", a.challenge)
}

func (a *Auth) HandleAuthWrite(req gatt.Request, data []byte) (status byte) {
	mac := hmac.New(sha256.New, a.Secret)
	mac.Write(a.challenge)
	expectedMac := mac.Sum(nil)
	equal := hmac.Equal(expectedMac, data)

	a.isAuthed = equal

	if a.isAuthed {
		fmt.Println("Successfully authenticated")
		return gatt.StatusSuccess
	} else {
		fmt.Println("Failed authentication")
		return gatt.StatusUnexpectedError
	}
}

func (a *Auth) Invalidate() {
	a.isAuthed = false
	fmt.Println("Invalidated authentication")
}
