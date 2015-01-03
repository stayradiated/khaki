package main

import (
	"fmt"
	"log"
	"time"

	"github.com/roylee17/gatt"
)

// services
var serviceUUID = gatt.MustParseUUID("54a64ddf-c756-4a1a-bf9d-14f2cac357ad")
var beaconUUID = gatt.MustParseUUID("a78d9129-b79a-400f-825e-b691661123eb")

// characteristics
var carUUID = gatt.MustParseUUID("fd1c6fcc-3ca5-48a9-97e9-37f81f5bd9c5")
var authUUID = gatt.MustParseUUID("66e01614-13d1-40d6-a34f-c5360ba57698")

// objects
var auth *Auth
var car *Car

// main starts up the BLE server
func main() {

	beaconData := NewBeacon(beaconUUID, 0, 0, 0x32)

	iBeacon := gatt.NewServer(
		gatt.Name("KhakiBeacon"),
		gatt.HCI("hci1"),
		gatt.AdvertisingPacket(beaconData.AdvertisingPacket()),
	)

	server := gatt.NewServer(
		gatt.Name("Khaki"),
		gatt.HCI("hci0"),
		gatt.Connect(HandleConnect),
		gatt.Disconnect(HandleDisconnect),
	)
	service := server.AddService(serviceUUID)

	// create auth instance
	auth = NewAuth([]byte("hunter2"))

	// auth characteristic
	authChar := service.AddCharacteristic(authUUID)
	authChar.HandleReadFunc(auth.HandleAuthRead)
	authChar.HandleWriteFunc(auth.HandleAuthWrite)

	// create car instance
	gpioPin, err := OpenGPIOPin()
	if err != nil {
		log.Println("Could not open GPIO pin")
	}
	car = NewCar(gpioPin, auth)

	// car characteristic
	carChar := service.AddCharacteristic(carUUID)
	carChar.HandleWriteFunc(car.HandleWrite)

	go log.Fatal(iBeacon.AdvertiseAndServe())
	log.Fatal(server.AdvertiseAndServe())
}

func HandleConnect(conn gatt.Conn) {
	fmt.Println("Got connection", conn)

	fmt.Println("You have 5 seconds...")

	go func() {
		time.Sleep(5 * time.Second)
		if !auth.IsAuthenticated() {
			fmt.Println("You have been disconnected")
			conn.Close()
		}
	}()
}

func HandleDisconnect(conn gatt.Conn) {
	fmt.Println("Lost connection", conn)
}
