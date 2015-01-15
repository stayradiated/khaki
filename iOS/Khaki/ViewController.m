//
//  ViewController.m
//  Khaki
//
//  Created by George Czabania on 3/01/15.
//  Copyright (c) 2015 Mintcode. All rights reserved.
//

#import "ViewController.h"

@interface ViewController ()

@end

@implementation ViewController

- (void)viewDidLoad {
    [super viewDidLoad];
    
    // Initiate CBCentralManager
    self.centralManager = [[CBCentralManager alloc] initWithDelegate:self queue:nil options:@{ CBCentralManagerOptionRestoreIdentifierKey: @"KhakiCentral" }];
    
    // Initiate CCLocationManager
    self.locationManager = [[CLLocationManager alloc] init];
    self.locationManager.delegate = self;
    if ([self.locationManager respondsToSelector:@selector(requestAlwaysAuthorization)]) {
        [self.locationManager requestAlwaysAuthorization];
    }
    
    // Initiate CLBeaconRegion
    NSLog(@"Looking for beacon: %@", KHAKI_BEACON_UUID);
    self.beaconRegion = [[CLBeaconRegion alloc] initWithProximityUUID:KHAKI_BEACON_UUID identifier:@"com.mintcode.beacon"];
    self.beaconRegion.notifyEntryStateOnDisplay = YES;
    [self.locationManager startMonitoringForRegion:self.beaconRegion];
}

- (void)didReceiveMemoryWarning {
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark - CBCentralManagerDelegate

// Called when the bluetooth chip changes state (i.e. powers on)
- (void)centralManagerDidUpdateState:(CBCentralManager *)central {
    
    if ([central state] == CBCentralManagerStatePoweredOn) {
        NSLog(@"CoreBluetooth BLE hardware is powered on");
        self.connectedLabel.text = @"Powered On";
        
        [self connectOrScan];
        
    } else {
        NSLog(@"CoreBluetooth BLE hardware is powered off");
        self.connectedLabel.text = @"Powered Off";
        
        self.peripheral = nil;
        self.carCharacteristic = nil;
        self.authCharacteristic = nil;
    }
}

- (void)centralManager:(CBCentralManager *)central willRestoreState:(NSDictionary *)state {
    // NSArray *scanServices = state[CBCentralManagerRestoredStateScanServicesKey];
    // NSArray *scanOptions = state[CBCentralManagerRestoredStateScanOptionsKey];
    
    NSLog(@"Restoring app!");
    
    NSArray *peripherals = state[CBCentralManagerRestoredStatePeripheralsKey];
    if ([peripherals count] > 0) {
        self.peripheral = peripherals[0];
        self.peripheral.delegate = self;
        self.peripheralUUID = self.peripheral.identifier;
    }
}

// Called when a bluetooth peripheral is discovered
- (void)centralManager:(CBCentralManager *)central didDiscoverPeripheral:(CBPeripheral *)peripheral advertisementData:(NSDictionary *)advertisementData RSSI:(NSNumber *)RSSI {
    NSLog(@"Found Peripheral");
    
    [self.centralManager stopScan];
    self.peripheral = peripheral;
    self.peripheral.delegate = self;
    [self.centralManager connectPeripheral:self.peripheral options:nil];
    
    self.connectedLabel.text = @"Discovered";
}

// Called when a bluetooth peripheral is connected to
- (void)centralManager:(CBCentralManager *)central didConnectPeripheral:(CBPeripheral *)peripheral {
    NSLog(@"Connected to peripheral");
    self.connectedLabel.text = @"Connected";
    
    self.peripheral = peripheral;
    self.peripheral.delegate = self;
    [self.peripheral discoverServices:@[KHAKI_SERVICE_UUID]];
}

// Called when a peripheral cannot be connected
- (void)centralManager:(CBCentralManager *)central didFailToConnectPeripheral:(CBPeripheral *)peripheral error:(NSError *)error {
    NSLog(@"Failed to connect to peripheral %@ (%@)", peripheral, error);
    self.connectedLabel.text = @"Failed To Connect";
}

// Called when a bluetooth peripheral is disconnected
- (void)centralManager:(CBCentralManager *)central didDisconnectPeripheral:(CBPeripheral *)peripheral error:(NSError *)error {
    NSLog(@"Disconnected peripheral %@ (%@)", peripheral, error);
    self.connectedLabel.text = [NSString stringWithFormat:@"Not Connected: %@", error];
    
    if (self.isInsideRegion) {
        [self connectOrScan];
    }
}

#pragma mark - CBPeripheralDelegate

// Called when the peripheral's services are discovered
- (void)peripheral:(CBPeripheral *)peripheral didDiscoverServices:(NSError *)error {
    for (CBService *service in peripheral.services) {
        NSLog(@"Discovered service: %@", service);
        
        if ([service.UUID isEqual:KHAKI_SERVICE_UUID]) {
            NSArray *characteristics = @[KHAKI_CAR_CHARACTERISTIC_UUID, KHAKI_AUTH_CHARACTERISTIC_UUID];
            [peripheral discoverCharacteristics:characteristics forService:service];
        }
    }
}

// Called when the characteristics for a service are discovered
- (void)peripheral:(CBPeripheral *)peripheral didDiscoverCharacteristicsForService:(CBService *)service error:(NSError *)error {
    if (! error) {
        for (CBCharacteristic *characteristic in service.characteristics) {
            
            // Car Characteristic
            if ([characteristic.UUID isEqual:KHAKI_CAR_CHARACTERISTIC_UUID]) {
                NSLog(@"Found Khaki Car characteristc");
                self.carCharacteristic = characteristic;
                [self.peripheral setNotifyValue:YES forCharacteristic:self.carCharacteristic];
                [self.peripheral readValueForCharacteristic:self.carCharacteristic];
            }
            
            // Auth Characteristic
            else if ([characteristic.UUID isEqual:KHAKI_AUTH_CHARACTERISTIC_UUID]) {
                NSLog(@"Found Khaki Auth characteristic");
                self.authCharacteristic = characteristic;
                [peripheral readValueForCharacteristic:characteristic];
            }
        }
    }
}

// Called when a value is read from a characteristic
- (void)peripheral:(CBPeripheral *)peripheral didUpdateValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    if (! error) {
        
        // Auth
        if ([characteristic.UUID isEqual:KHAKI_AUTH_CHARACTERISTIC_UUID]) {
            [self authenticate:characteristic];
            
        // Car
        } else if ([characteristic.UUID isEqual:KHAKI_CAR_CHARACTERISTIC_UUID]) {
            [self readCarStatus:characteristic];
        }
    }
}

#pragma mark - CLLocationManagerDelegate

// Called when we move inside/outside the region, or when the user wakes up the phone
- (void)locationManager:(CLLocationManager *)manager didDetermineState:(CLRegionState)state forRegion:(CLRegion *)region {
    if (state == CLRegionStateInside) {
        self.isInsideRegion = YES;
        self.iBeaconLabel.text = @"INSIDE";
        [self connectOrScan];
    }
    else if (state == CLRegionStateOutside) {
        NSLog(@"OUTSIDE");
        self.isInsideRegion = NO;
        self.iBeaconLabel.text = @"OUTSIDE";
        [self.locationManager stopRangingBeaconsInRegion:self.beaconRegion];
        [self.centralManager stopScan];
        
        self.connectedLabel.text = @"Waiting until in range...";
    }
    else {
        NSLog(@"OTHER");
        self.iBeaconLabel.text = @"OTHER";
    }
}

// Called when we get a range update
- (void)locationManager:(CLLocationManager *)manager didRangeBeacons:(NSArray *)beacons inRegion:(CLBeaconRegion *)region {
    for (CLBeacon *beacon in beacons) {
        self.iBeaconLabel.text = [NSString stringWithFormat:@"Prox: %ld -- RSSI: %ld", (long) beacon.proximity, (long) beacon.rssi];
        
        // if (beacon.proximity <= CLProximityNear) {
        if (beacon.rssi >= -50 && beacon.rssi < 0) {
            if (! self.isUnlocked) {
                [self unlock];
            }
        } else {
            if (self.isUnlocked) {
                [self lock];
            }
        }
    }
}

#pragma mark - Buttons

- (IBAction)tapLockButton:(id)sender {
    NSLog(@"Tapping Lock Button");
    [self lock];
}

- (IBAction)tapUnlockButton:(id)sender {
    NSLog(@"Tapping Unlock Button");
    [self unlock];
}

#pragma mark - Khaki methods

- (void)connectOrScan {
    
    NSLog(@"Connect/Scan started");
    
    if (self.peripheral && self.peripheral.state == CBPeripheralStateConnected) {
        
        NSLog(@"Already connected to a peripheral");
        
        // Check if we have discovered the service
        NSUInteger serviceIdx = [self.peripheral.services indexOfObjectPassingTest:^BOOL(CBService *obj, NSUInteger idx, BOOL *stop) {
            return [obj.UUID isEqual:KHAKI_SERVICE_UUID];
        }];
        
        if (serviceIdx == NSNotFound) {
            // We haven't discovered all the services yet
            NSLog(@"Have not yet discovered services");
            [self.peripheral discoverServices:@[KHAKI_SERVICE_UUID]];
            return;
        }
        
        CBService *service = self.peripheral.services[serviceIdx];
        
        // Check if we have discovered the car characteristic
        NSUInteger carCharIdx = [service.characteristics indexOfObjectPassingTest:^BOOL(CBCharacteristic *obj, NSUInteger idx, BOOL *stop) {
            return [obj.UUID isEqual:KHAKI_CAR_CHARACTERISTIC_UUID];
        }];
        
        // Check if we have discovered the car characteristic
        NSUInteger authCharIdx = [service.characteristics indexOfObjectPassingTest:^BOOL(CBCharacteristic *obj, NSUInteger idx, BOOL *stop) {
            return [obj.UUID isEqual:KHAKI_AUTH_CHARACTERISTIC_UUID];
        }];
        
        // Build an array of services that we haven't got yet
        NSMutableArray *services = [[NSMutableArray alloc] init];
        
        if (carCharIdx == NSNotFound) {
            NSLog(@"Have not yet discovered car characteristic");
            [services addObject:KHAKI_CAR_CHARACTERISTIC_UUID];
        } else {
            self.carCharacteristic = service.characteristics[carCharIdx];
            
            // Make sure we are subscribed
            if (! self.carCharacteristic.isNotifying) {
                [self.peripheral setNotifyValue:YES forCharacteristic:self.carCharacteristic];
            }
        }
        
        if (authCharIdx == NSNotFound) {
            NSLog(@"Have not yet discovered auth characteristic");
            [services addObject:KHAKI_AUTH_CHARACTERISTIC_UUID];
        } else {
            self.authCharacteristic = service.characteristics[authCharIdx];
        }
        
        if ([services count] > 0) {
            // We haven't discovered all the characteristic yet!
            [self.peripheral discoverCharacteristics:services forService:service];
            return;
        }
        
        NSLog(@"Already discovered services successfully");
        return;
    }
    
    // If we have previously connected to a device, then we will have it's UUID.
    if (self.peripheralUUID) {
        NSLog(@"Using self.peripheralUUID: %@", self.peripheralUUID);
        
        NSArray *peripherals = [self.centralManager retrievePeripheralsWithIdentifiers:@[self.peripheralUUID]];
        if ([peripherals count] > 0) {
            
            NSLog(@"Found a peripheral, trying to connect to it");
            [self.centralManager connectPeripheral:peripherals[0] options:nil];
            
            // TODO: If we can't connect, then we need to do a scan
            return;
        }
        
    }
    
    // Check if are are already connected to it
    NSLog(@"Checking connected peripherals");
    NSArray *peripherals = [self.centralManager retrieveConnectedPeripheralsWithServices:@[KHAKI_SERVICE_UUID]];
    if ([peripherals count] > 0) {
        NSLog(@"Found a connected peripheral");
        [self.centralManager connectPeripheral:peripherals[0] options:nil];
        return;
    }
    
    // Start scanning for services
    NSLog(@"Running a scan");
    self.connectedLabel.text = @"Scanning...";
    [self.centralManager scanForPeripheralsWithServices:@[KHAKI_SERVICE_UUID] options:nil];
}

- (void)authenticate:(CBCharacteristic *)characteristic {
    
    // Hash bytes using HMAC SHA256
    const char *key = [@"hunter2" cStringUsingEncoding:NSASCIIStringEncoding];
    const char *data = [characteristic.value bytes];
    const long dataLength = [characteristic.value length] / sizeof(char);
    
    NSLog(@"Direct data: %@", characteristic.value);
    NSLog(@"Data: %@", [NSData dataWithBytes:data length:dataLength]);
    
    unsigned char cHMAC[CC_SHA256_DIGEST_LENGTH];
    CCHmac(kCCHmacAlgSHA256, key, strlen(key), data, dataLength, cHMAC);
    NSData *response = [[NSData alloc] initWithBytes:cHMAC length:sizeof(cHMAC)];
    
    NSLog(@"HMAC: %@", response);
    
    [self.peripheral writeValue:response forCharacteristic:characteristic type:CBCharacteristicWriteWithoutResponse];
}

- (void)unlock {
    NSLog(@"Unlocking car");
    unsigned char status = 0 | KHAKI_UNLOCKED;
    if ([self writeCarStatus:status]) {
        self.isUnlocked = true;
        self.carStatusLabel.text = @"Unlocked";
    }
}

- (void)lock {
    NSLog(@"Locking car");
    unsigned char status = 0;
    if ([self writeCarStatus:status]) {
        self.isUnlocked = false;
        self.carStatusLabel.text = @"Locked";
    }
}

- (BOOL)writeCarStatus:(unsigned char)status {
    const unsigned char bytes[] = {status};
    NSData *data = [NSData dataWithBytes:bytes length:sizeof(bytes)];
    
    if (self.peripheral == nil || self.carCharacteristic == nil) {
        NSLog(@"Not yet connected...");
        return NO;
    }
    
    NSLog(@"Writing %@", data);
    [self.peripheral writeValue:data forCharacteristic:self.carCharacteristic type:CBCharacteristicWriteWithoutResponse];
    
    return YES;
}

- (BOOL)readCarStatus:(CBCharacteristic *)characteristic {
    const char *bytes = [characteristic.value bytes];
    int firstByte = bytes[0];
    
    NSLog(@"Reading Car Status");
    
    if ((firstByte & KHAKI_UNLOCKED) != 0) {
        self.isUnlocked = true;
    } else {
        self.isUnlocked = false;
    }
    
    if ((firstByte & KHAKI_NOTIFYING) != 0) {
        if (! self.isRanging) {
            self.isRanging = true;
            
            NSLog(@"Start Ranging");
            [self.locationManager startRangingBeaconsInRegion:self.beaconRegion];
            self.iBeaconLabel.text = @"Started Ranging";

        }
        
    } else {
        if (self.isRanging) {
            self.isRanging = false;
            
            NSLog(@"Stop Ranging");
            [self.locationManager stopRangingBeaconsInRegion:self.beaconRegion];
            self.iBeaconLabel.text = @"Stopped Ranging";
        }
    }
    
    NSLog(@"Car is locked: %x", self.isUnlocked);
    
    return self.isUnlocked;
}

@end
