package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/paypal/gatt"
)

var serviceUUID = gatt.MustParseUUID("54a64ddf-c756-4a1a-bf9d-14f2cac357ad")
var carUUID = gatt.MustParseUUID("fd1c6fcc-3ca5-48a9-97e9-37f81f5bd9c5")
var authUUID = gatt.MustParseUUID("66e01614-13d1-40d6-a34f-c5360ba57698")

func main() {
	server := gatt.NewServer(
		gatt.Name("Khaki"),
		gatt.AdvertiseServices([]gatt.UUID{serviceUUID}),
		gatt.HCI("hci0"),
		gatt.MaxConnections(1),
		gatt.Connect(HandleConnect),
		gatt.Disconnect(HandleDisconnect),
	)
	service := server.AddService(serviceUUID)

	gpioPin, err := OpenGPIOPin()
	if err != nil {
		panic(errors.New("Could not open GPIO pin"))
	}
	car := &Car{Pin: gpioPin}

	carChar := service.AddCharacteristic(carUUID)
	carChar.HandleWriteFunc(car.HandleWrite)

	authChar := service.AddCharacteristic(authUUID)
	authChar.HandleReadFunc(HandleAuthRead)
	authChar.HandleWriteFunc(HandleAuthWrite)

	log.Fatal(server.AdvertiseAndServe())
}

func HandleConnect(conn gatt.Conn) {
	fmt.Println("Got connection")
	fmt.Println(conn)
}

func HandleDisconnect(conn gatt.Conn) {
	fmt.Println("Lost connection")
	fmt.Println(conn)
}
