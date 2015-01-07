package main

import (
	"encoding/hex"

	"github.com/paypal/gatt"
)

// Based on https://github.com/izqui/beacon

type iBeaconConfig struct {
	UUID  gatt.UUID
	Major uint16
	Minor uint16
	Power byte
}

func iBeaconPacket(c *iBeaconConfig) []byte {
	uuid, err := hex.DecodeString(c.UUID.String())
	if err != nil {
		panic("Could not parse UUID")
	}

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
	packet = append(packet, uuid...)

	packet = append(packet,
		// iBeacon Major
		uint8(c.Major>>8),
		uint8(c.Major&0xff),

		// iBeacon Minor
		uint8(c.Minor>>8),
		uint8(c.Minor&0xff),

		// iBeacon Power
		c.Power,
	)

	return packet
}
