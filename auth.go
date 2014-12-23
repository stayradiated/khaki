package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/dustin/randbo"
	"github.com/paypal/gatt"
)

var lastChallenge []byte = make([]byte, 16)
var key []byte = []byte("hunter2")
var authed = true

func HandleAuthRead(resp gatt.ReadResponseWriter, req *gatt.ReadRequest) {
	randbo.New().Read(lastChallenge)
	resp.Write(lastChallenge)
	fmt.Println(lastChallenge)
}

func HandleAuthWrite(req gatt.Request, data []byte) (status byte) {
	mac := hmac.New(sha256.New, key)
	mac.Write(lastChallenge)
	expectedMac := mac.Sum(nil)
	equal := hmac.Equal(expectedMac, data)

	authed = equal

	fmt.Println("data:", data)
	fmt.Println("hmac:", expectedMac)
	fmt.Println("equal:", equal)

	return gatt.StatusSuccess
}
