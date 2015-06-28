package main

import "github.com/paypal/gatt"

var (
	// beacons
	beaconUUID = gatt.MustParseUUID("a78d9129-b79a-400f-825e-b691661123eb")

	// services
	serviceUUID = gatt.MustParseUUID("54a64ddf-c756-4a1a-bf9d-14f2cac357ad")

	// characteristics
	carUUID  = gatt.MustParseUUID("fd1c6fcc-3ca5-48a9-97e9-37f81f5bd9c5")
	authUUID = gatt.MustParseUUID("66e01614-13d1-40d6-a34f-c5360ba57698")
)
