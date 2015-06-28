package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"

	"github.com/dustin/randbo"
	"github.com/paypal/gatt"
)

// Auth authenticates a connection using HMAC-SHA256
type Auth struct {
	mu        sync.Mutex
	isAuthed  bool
	public    bool // if true, then authentication can be bypassed
	random    io.Reader
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
		random:    randbo.New(),
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

// NextChallenege creates a new challenge
func (a *Auth) NextChallenge() []byte {
	buf := make([]byte, 16)
	a.random.Read(buf)
	a.mu.Lock()
	a.challenge = buf
	a.mu.Unlock()
	return buf
}

// TestChallenge checks the input against the HMAC-SHA256 of the challenge
func (a *Auth) TestChallenge(data []byte) (status byte) {
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
