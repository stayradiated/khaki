package main

import "testing"

func TestAuth(t *testing.T) {

	auth := NewAuth([]byte("secret"), false)

	// should be logged out by default
	if auth.IsAuthenticated() != false {
		t.Fatal("Auth is not logged out by default")
	}

}
