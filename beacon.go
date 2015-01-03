package main

import (
	"encoding/hex"

	"github.com/paypal/gatt"
)

// Based on https://github.com/izqui/beacon

type Beacon struct {
	UUID  []byte
	Major uint16
	Minor uint16
	Power byte
}

func NewBeacon(uuid gatt.UUID, major uint16, minor uint16, power byte) *Beacon {
	uuidBytes, err := hex.DecodeString(uuid.String())
	if err != nil {
		panic("Could not parse UUID")
	}

	return &Beacon{
		UUID:  uuidBytes,
		Major: major,
		Minor: minor,
		Power: power,
	}
}

func (b Beacon) AdvertisingPacket() []byte {

	packet := []byte{
		0x02, // Number of bytes that follow in first AD structure
		0x01, // Flags AD type
		0x1A, // Flag value 0x1A = 0001 1010
		0x1A, // Number of bytes that follow in second advertising structure
		0xFF, // Manafacturer specific data advertising type

		// Apple company identifier
		0x4c,
		0x00,

		// iBeacon identifier
		0x02,
		0x15,
	}

	// iBeacon UUID
	packet = append(packet, b.UUID...)

	packet = append(packet,
		// iBeacon Major
		uint8(b.Major>>8),
		uint8(b.Major&0xff),

		// iBeacon Minor
		uint8(b.Minor>>8),
		uint8(b.Minor&0xff),

		// iBeacon Power
		b.Power,
	)

	return packet
}
