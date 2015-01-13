package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecheney/gpio/rpi"
	"github.com/paypal/gatt"
)

var (
	// services
	serviceUUID = gatt.MustParseUUID("54a64ddf-c756-4a1a-bf9d-14f2cac357ad")

	// characteristics
	carUUID  = gatt.MustParseUUID("fd1c6fcc-3ca5-48a9-97e9-37f81f5bd9c5")
	authUUID = gatt.MustParseUUID("66e01614-13d1-40d6-a34f-c5360ba57698")

	// beacons
	beaconUUID = gatt.MustParseUUID("a78d9129-b79a-400f-825e-b691661123eb")
)

type PeripheralConfig struct {
	Secret []byte
	Public bool
}

type Peripheral struct {
	Beacon *gatt.Server
	Server *gatt.Server

	Car       *Car
	Auth      *Auth
	StatusLED *StatusLED
	Sensor    *Sensor
}

func NewPeripheral(c *PeripheralConfig) *Peripheral {
	p := new(Peripheral)
	p.Init(c)
	return p
}

// main starts up the BLE server
func (p *Peripheral) Init(c *PeripheralConfig) {

	p.Beacon = gatt.NewServer(
		gatt.Name("KhakiBeacon"),
		gatt.HCI(1),
		gatt.AdvertisingPacket(iBeaconPacket(&iBeaconConfig{
			UUID:  beaconUUID,
			Major: 0,
			Minor: 0,
			Power: 0xCD,
		})),
	)

	p.Server = gatt.NewServer(
		gatt.Name("Khaki"),
		gatt.HCI(0),
		gatt.Connect(p.handleConnect),
		gatt.Disconnect(p.handleDisconnect),
	)
	service := p.Server.AddService(serviceUUID)

	// create status instance
	gpioPin24, err := OpenPinForOutput(rpi.GPIO24)
	if err != nil {
		log.Println("Could not open GPIO pin 24")
	}
	p.StatusLED = &StatusLED{
		Pin: gpioPin24,
	}

	// create auth instance
	p.Auth = NewAuth(c.Secret, c.Public)

	// auth characteristic
	authChar := service.AddCharacteristic(authUUID)
	authChar.HandleReadFunc(p.Auth.HandleRead)
	authChar.HandleWriteFunc(p.Auth.HandleWrite)

	// create car instance
	gpioPin25, err := OpenPinForOutput(rpi.GPIO25)
	if err != nil {
		log.Println("Could not open GPIO pin 25")
	}
	p.Car = NewCar(gpioPin25, p.Auth)

	// car characteristic
	carChar := service.AddCharacteristic(carUUID)
	carChar.HandleReadFunc(p.Car.HandleRead)
	carChar.HandleWriteFunc(p.Car.HandleWrite)
	carChar.HandleNotifyFunc(p.Car.HandleNotify)

	// sensor
	gpioPin23, err := OpenPinForInput(rpi.GPIO23)
	if err != nil {
		log.Println("Could not open GPIO pin 23")
	}
	p.Sensor = &Sensor{
		Pin: gpioPin23,
	}
	p.Sensor.Watch()
}

// Start starts running the BLE servers
func (p *Peripheral) Start() {
	go func() {
		err := p.Beacon.AdvertiseAndServe()
		if err != nil {
			log.Printf("Error with iBeacon: %s\n", err)
		}
	}()

	go func() {
		log.Fatal(p.Server.AdvertiseAndServe())
	}()

	select {}
}

// HandleConnect is called when a central connects
func (p *Peripheral) handleConnect(conn gatt.Conn) {
	fmt.Println("Got connection", conn)
	p.StatusLED.Update(true)

	fmt.Println("You have 5 seconds...")

	go func() {
		time.Sleep(5 * time.Second)
		if !p.Auth.IsAuthenticated() {
			fmt.Println("You have been disconnected")
			conn.Close()
		}
	}()
}

// HandleDisconnect is called when a connection is lost
func (p *Peripheral) handleDisconnect(conn gatt.Conn) {
	fmt.Println("Lost connection", conn)
	p.Car.Lock()
	p.Auth.Invalidate()
	p.StatusLED.Update(false)
}
