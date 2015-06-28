package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecheney/gpio/rpi"
	"github.com/paypal/gatt"
	"github.com/stayradiated/shifty"
)

var (
	// devices
	beaconDeviceID = 0
	serverDeviceID = 1
)

type Peripheral struct {
	beacon         gatt.Device
	server         gatt.Device
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
	var err error

	p.beacon, err = gatt.NewDevice(gatt.LnxDeviceID(beaconDeviceID, false))
	if err != nil {
		log.Fatalf("Could not open HCI Device %d", beaconDeviceID)
	}

	p.server, err = gatt.NewDevice(gatt.LnxDeviceID(serverDeviceID, false))
	if err != nil {
		log.Fatalf("Could not open HCI Device %d", serverDeviceID)
	}

	p.InitGPIO()
	p.InitBeacon(p.beacon)
	p.InitServer(p.server)
}

// InitBeacon starts an iBeacon peripheral
func (p *Peripheral) InitBeacon(d gatt.Device) {

	d.Init(func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			err := d.AdvertiseIBeacon(
				beaconUUID, // uuid
				0,          // major
				0,          // minor
				-77,        // power
			)
			if err != nil {
				log.Fatal(err)
			}
		default:
		}
	})

	/*
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
	*/

}

func (p *Peripheral) InitServer(d gatt.Device) {
	d.Handle(
		gatt.CentralConnected(p.handleConnect),
		gatt.CentralDisconnected(p.handleDisconnect),
	)

	d.Init(func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			d.AddService(NewKhakiService(p.auth, p.car))
			d.AdvertiseNameAndServices("Khaki", []gatt.UUID{serviceUUID})
		default:
		}
	})
}

// HandleConnect is called when a central connects
func (p *Peripheral) handleConnect(c gatt.Central) {
	log.Println("Got connection", c.ID())
	p.bluetoothLED.Update(true)

	log.Println("You have 5 seconds...")

	go func() {
		time.Sleep(5 * time.Second)
		if !p.auth.IsAuthenticated() {
			log.Println("You have been disconnected")
			c.Close()
		}
	}()
}

// HandleDisconnect is called when a connection is lost
func (p *Peripheral) handleDisconnect(c gatt.Central) {
	log.Println("Lost connection", c.ID())
	p.car.Lock()
	p.car.Reset()
	p.auth.Invalidate()
	p.bluetoothLED.Update(false)
}

func (p *Peripheral) InitGPIO() {

	// set up shift register
	s := &shifty.ShiftRegister{
		DataPin:  MustOpenPinForOutput(rpi.GPIO17),
		LatchPin: MustOpenPinForOutput(rpi.GPIO27),
		ClockPin: MustOpenPinForOutput(rpi.GPIO22),
		MaxPins:  16,
	}

	p.piLED = NewStatusLED(s.Pin(0), 2*time.Second, true)
	p.bluetoothLED = NewStatusLED(s.Pin(1), 0, false)
	p.powerSavingLED = NewStatusLED(s.Pin(2), 0, false)

	// create car instance
	p.car = NewCar(s.Pin(3), p.auth)

	// Acc Power Sensor
	accPin, err := OpenPinForInput(rpi.GPIO23)
	if err != nil {
		log.Println("Could not open GPIO pin 23")
	}

	// sensor
	p.accPowerSensor = &Sensor{
		Pin: accPin,
		HandleChange: func(sensor bool) {
			p.car.ToggleNotifications(sensor)
			p.powerSavingLED.Update(sensor)
		},
	}
	p.accPowerSensor.Init()

}
