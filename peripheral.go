package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecheney/gpio/rpi"
	"github.com/paypal/gatt"
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
var status *Status

type PeripheralConfig struct {
	Secret string
	Public bool
}

// main starts up the BLE server
func StartPeripheral(c *PeripheralConfig) {

	iBeacon := gatt.NewServer(
		gatt.Name("KhakiBeacon"),
		gatt.HCI(1),
		gatt.AdvertisingPacket(iBeaconPacket(&iBeaconConfig{
			UUID:  beaconUUID,
			Major: 0,
			Minor: 0,
			Power: 0xCD,
		})),
	)

	server := gatt.NewServer(
		gatt.Name("Khaki"),
		gatt.HCI(0),
		gatt.Connect(HandleConnect),
		gatt.Disconnect(HandleDisconnect),
	)
	service := server.AddService(serviceUUID)

	// create status instance
	gpioPin24, err := OpenGPIOPin(rpi.GPIO24)
	if err != nil {
		log.Println("Could not open GPIO pin 24")
	}
	status = &Status{
		Connected: false,
		Pin:       gpioPin24,
	}

	// create auth instance
	auth = NewAuth(&AuthConfig{
		Secret: []byte(c.Secret),
		Public: c.Public,
	})

	// auth characteristic
	authChar := service.AddCharacteristic(authUUID)
	authChar.HandleReadFunc(auth.HandleAuthRead)
	authChar.HandleWriteFunc(auth.HandleAuthWrite)

	// create car instance
	gpioPin25, err := OpenGPIOPin(rpi.GPIO25)
	if err != nil {
		log.Println("Could not open GPIO pin 25")
	}
	car = NewCar(gpioPin25, auth)

	// car characteristic
	carChar := service.AddCharacteristic(carUUID)
	carChar.HandleReadFunc(car.HandleRead)
	carChar.HandleWriteFunc(car.HandleWrite)
	carChar.HandleNotifyFunc(car.HandleNotify)

	go func() {
		err := iBeacon.AdvertiseAndServe()
		if err != nil {
			log.Printf("Error with iBeacon: %s\n", err)
		}
	}()

	go func() {
		log.Fatal(server.AdvertiseAndServe())
	}()

	select {}
}

// HandleConnect is called when a central connects
func HandleConnect(conn gatt.Conn) {
	fmt.Println("Got connection", conn)
	status.Update(true)

	fmt.Println("You have 5 seconds...")

	go func() {
		time.Sleep(5 * time.Second)
		if !auth.IsAuthenticated() {
			fmt.Println("You have been disconnected")
			conn.Close()
		}
	}()
}

// HandleDisconnect is called when a connection is lost
func HandleDisconnect(conn gatt.Conn) {
	fmt.Println("Lost connection", conn)
	car.Lock()
	auth.Invalidate()
	status.Update(false)
}
