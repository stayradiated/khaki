package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/dustin/randbo"
	"github.com/paypal/gatt"
)

// Auth authenticates a connection using HMAC-SHA256
type Auth struct {
	mu        sync.Mutex
	isAuthed  bool
	public    bool // if true, then authentication can be bypassed
	challenge []byte
	Secret    []byte
}

// NewAuth creates a new Auth instance
func NewAuth(secret []byte, public bool) *Auth {
	return &Auth{
		challenge: make([]byte, 16),
		isAuthed:  false,
		public:    public,
		Secret:    secret,
	}
}

// IsAuthenticated reports current authentication status
func (a Auth) IsAuthenticated() bool {
	var isAuthed bool
	if a.public {
		isAuthed = true
	} else {
		a.mu.Lock()
		isAuthed = a.isAuthed
		a.mu.Unlock()
	}
	return isAuthed
}

// HandleRead creates a new challenge
func (a *Auth) HandleRead(resp gatt.ReadResponseWriter, req *gatt.ReadRequest) {
	randbo.New().Read(a.challenge)
	resp.Write(a.challenge)
	fmt.Printf("Creating new challenge: %x\n", a.challenge)
}

// HandleWrite checks the input against the HMAC-SHA256 of the challenge
func (a *Auth) HandleWrite(req gatt.Request, data []byte) (status byte) {
	mac := hmac.New(sha256.New, a.Secret)
	mac.Write(a.challenge)
	expectedMac := mac.Sum(nil)
	equal := hmac.Equal(expectedMac, data)

	a.mu.Lock()
	a.isAuthed = equal
	a.mu.Unlock()

	if equal {
		fmt.Println("Successfully authenticated")
		return gatt.StatusSuccess
	} else {
		fmt.Println("Failed authentication")
		return gatt.StatusUnexpectedError
	}
}

// Invalidate deauthenticates the session
func (a *Auth) Invalidate() {
	a.mu.Lock()
	a.isAuthed = false
	a.mu.Unlock()

	fmt.Println("Invalidated authentication")
}
