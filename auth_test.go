package main

import (
	"bytes"
	"testing"
)

func TestAuth(t *testing.T) {

	auth := NewAuth([]byte("secret"), false)

	// should be logged out by default
	if auth.IsAuthenticated() != false {
		t.Fatal("Auth is not logged out by default")
	}

	// should return a random challenge
	challengeA := auth.NextChallenge()
	challengeB := auth.NextChallenge()

	if bytes.Equal(challengeA, challengeB) {
		t.Logf("%x\n", challengeA)
		t.Logf("%x\n", challengeB)
		t.Fatal("Returned the same challenge twice")
	}

	if len(challengeA) != 16 {
		t.Fatal("Challenge is incorrect length", len(challengeA))
	}

}
