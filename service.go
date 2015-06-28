package main

import "github.com/paypal/gatt"

func NewKhakiService(auth *Auth, car *Car) *gatt.Service {
	s := gatt.NewService(serviceUUID)

	authChar := s.AddCharacteristic(authUUID)
	authChar.HandleReadFunc(func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
		rsp.Write(auth.NextChallenge())
	})
	authChar.HandleWriteFunc(func(req gatt.Request, data []byte) (status byte) {
		return auth.TestChallenge(data)
	})

	carChar := s.AddCharacteristic(carUUID)
	carChar.HandleReadFunc(func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
		rsp.Write(car.Status())
	})
	carChar.HandleWriteFunc(func(req gatt.Request, data []byte) (status byte) {
		return car.Write(data)
	})
	carChar.HandleNotifyFunc(func(req gatt.Request, n gatt.Notifier) {
		car.Notify(n)
	})

	return s
}
