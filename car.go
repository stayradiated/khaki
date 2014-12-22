package main

import "github.com/paypal/gatt"

func HandleCarWrite(r gatt.Request, data []byte) (status byte) {
	if !authed {
		return gatt.StatusUnexpectedError
	}
	return gatt.StatusSuccess
}
