package main

import (
	"log"
	"time"

	"github.com/davecheney/gpio/rpi"
	"github.com/paypal/gatt"
	"github.com/stayradiated/shifty"
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

type Peripheral struct {
	beacon         *gatt.Server
	server         *gatt.Server
	car            *Car
	auth           *Auth
	piLED          *StatusLED
	bluetoothLED   *StatusLED
	powerSavingLED *StatusLED
	accPowerSensor *Sensor
	doorLockSensor *Sensor
}

func NewPeripheral(auth *Auth) *Peripheral {
	p := new(Peripheral)
	p.auth = auth
	p.Init()
	return p
}

// main starts up the BLE server
func (p *Peripheral) Init() {

	p.beacon = gatt.NewServer(
		gatt.Name("KhakiBeacon"),
		gatt.HCI(1),
		gatt.AdvertisingPacket(iBeaconPacket(&iBeaconConfig{
			UUID:  beaconUUID,
			Major: 0,
			Minor: 0,
			Power: 0xCD,
		})),
	)

	p.server = gatt.NewServer(
		gatt.Name("Khaki"),
		gatt.HCI(0),
		gatt.Connect(p.handleConnect),
		gatt.Disconnect(p.handleDisconnect),
	)
	service := p.server.AddService(serviceUUID)

	// set up shift register
	s := &shifty.ShiftRegister{
		DataPin:  MustOpenPinForOutput(rpi.GPIO17),
		LatchPin: MustOpenPinForOutput(rpi.GPIO27),
		ClockPin: MustOpenPinForOutput(rpi.GPIO22),
		MaxPins:  16,
	}

	// Acc Power Sensor
	gpioPin23, err := OpenPinForInput(rpi.GPIO23)
	if err != nil {
		log.Println("Could not open GPIO pin 23")
	}

	p.piLED = &StatusLED{
		Pin:   s.Pin(0),
		Blink: 2 * time.Second,
	}

	log.Println("LED Should be turning on maybe?")
	p.piLED.Update(true)

	p.bluetoothLED = &StatusLED{
		Pin:   s.Pin(1),
		Blink: 0,
	}

	p.powerSavingLED = &StatusLED{
		Pin:   s.Pin(2),
		Blink: 0,
	}

	// auth characteristic
	authChar := service.AddCharacteristic(authUUID)
	authChar.HandleReadFunc(p.auth.HandleRead)
	authChar.HandleWriteFunc(p.auth.HandleWrite)

	// create car instance
	p.car = NewCar(s.Pin(3), p.auth)

	// car characteristic
	carChar := service.AddCharacteristic(carUUID)
	carChar.HandleReadFunc(p.car.HandleRead)
	carChar.HandleWriteFunc(p.car.HandleWrite)
	carChar.HandleNotifyFunc(p.car.HandleNotify)

	// sensor
	p.accPowerSensor = &Sensor{
		Pin: gpioPin23,
		HandleChange: func(sensor bool) {
			p.car.ToggleNotifications(sensor)
			p.powerSavingLED.Update(sensor)
		},
	}
	p.accPowerSensor.Init()
}

// Start starts running the BLE servers
func (p *Peripheral) Start() {
	go func() {
		err := p.beacon.AdvertiseAndServe()
		if err != nil {
			log.Printf("Error with iBeacon: %s\n", err)
		}
	}()

	go func() {
		log.Fatal(p.server.AdvertiseAndServe())
	}()

	select {}
}

// HandleConnect is called when a central connects
func (p *Peripheral) handleConnect(conn gatt.Conn) {
	log.Println("Got connection", conn)
	p.bluetoothLED.Update(true)

	log.Println("You have 5 seconds...")

	go func() {
		time.Sleep(5 * time.Second)
		if !p.auth.IsAuthenticated() {
			log.Println("You have been disconnected")
			conn.Close()
		}
	}()
}

// HandleDisconnect is called when a connection is lost
func (p *Peripheral) handleDisconnect(conn gatt.Conn) {
	log.Println("Lost connection", conn)
	p.car.Lock()
	p.car.Reset()
	p.auth.Invalidate()
	p.bluetoothLED.Update(false)
}
