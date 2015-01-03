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
    self.centralManager = [[CBCentralManager alloc] initWithDelegate:self queue:nil];
}

- (void)didReceiveMemoryWarning {
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark - CBCentralManagerDelegate

// Called when the bluetooth chip changes state (i.e. powers on)
- (void)centralManagerDidUpdateState:(CBCentralManager *)central {
    switch ([central state]) {
        case CBCentralManagerStatePoweredOn: {
            NSLog(@"CoreBluetooth BLE hardware is powered on");
            NSArray *services = @[KHAKI_SERVICE_UUID];
            [self.centralManager scanForPeripheralsWithServices:services options:nil];
            NSLog(@"Scanning for Khaki service: %@", services);
            break;
        }
        case CBCentralManagerStatePoweredOff: {
            NSLog(@"CoreBluetooth BLE hardware is powered off");
            break;
        }
        case CBCentralManagerStateUnauthorized: {
            NSLog(@"CoreBluetooth BLE hardware is unauthorized");
            break;
        }
        case CBCentralManagerStateUnsupported: {
            NSLog(@"CoreBluetooth BLE hardware is unsupported");
            break;
        }
        case CBCentralManagerStateResetting: {
            NSLog(@"CoreBluetooth BLE hardware is resetting");
            break;
        }
        case CBCentralManagerStateUnknown: {
            NSLog(@"CoreBluetooth BLE hardware is in an unknown state");
            break;
        }
    }
}

// Called when a bluetooth peripheral is discovered
- (void)centralManager:(CBCentralManager *)central didDiscoverPeripheral:(CBPeripheral *)peripheral advertisementData:(NSDictionary *)advertisementData RSSI:(NSNumber *)RSSI {
    
    NSLog(@"Found Peripheral: %@", peripheral);
    NSLog(@"Advertisement: %@", advertisementData);
    
    [self.centralManager stopScan];
    self.peripheral = peripheral;
    self.peripheral.delegate = self;
    [self.centralManager connectPeripheral:self.peripheral options:nil];
}

// Called when a bluetooth peripheral is connected to
- (void)centralManager:(CBCentralManager *)central didConnectPeripheral:(CBPeripheral *)peripheral {
    [peripheral setDelegate:self];
    NSArray *services = @[KHAKI_SERVICE_UUID];
    [peripheral discoverServices:services];
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
    for (CBCharacteristic *characteristic in service.characteristics) {
        
        // Car Characteristic
        if ([characteristic.UUID isEqual:KHAKI_CAR_CHARACTERISTIC_UUID]) {
            NSLog(@"Found Khaki Car characteristc");
            self.carCharacteristic = characteristic;
        }
        
        // Auth Characteristic
        else if ([characteristic.UUID isEqual:KHAKI_AUTH_CHARACTERISTIC_UUID]) {
            NSLog(@"Found Khaki Auth characteristic");
            self.authCharacteristic = characteristic;
            [peripheral readValueForCharacteristic:characteristic];
        }
    }
}

// Called when a value is read from a characteristic
- (void)peripheral:(CBPeripheral *)peripheral didUpdateValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    
    if ([characteristic.UUID isEqual:KHAKI_AUTH_CHARACTERISTIC_UUID]) {
        [self authenticate:characteristic];
    }
    
}

#pragma mark - Buttons

- (IBAction)tapUnlockButton:(id)sender {
    NSLog(@"Tapping button");
    [self unlock];
}

#pragma mark - Khaki methods

- (void)authenticate:(CBCharacteristic *)characteristic {
    
    // Hash bytes using HMAC SHA256
    const char *key = [@"hunter2" cStringUsingEncoding:NSASCIIStringEncoding];
    const char *data = [[characteristic value] bytes];
    const long dataLength = [[characteristic value] length] / sizeof(char);
    
    NSLog(@"Direct data: %@", [characteristic value]);
    NSLog(@"Data: %@", [NSData dataWithBytes:data length:dataLength]);
    
    unsigned char cHMAC[CC_SHA256_DIGEST_LENGTH];
    CCHmac(kCCHmacAlgSHA256, key, strlen(key), data, dataLength, cHMAC);
    NSData *response = [[NSData alloc] initWithBytes:cHMAC length:sizeof(cHMAC)];
    
    NSLog(@"HMAC: %@", response);
    
    [self.peripheral writeValue:response forCharacteristic:characteristic type:CBCharacteristicWriteWithoutResponse];
}

- (void)unlock {
    const unsigned char bytes[] = {2};
    NSData *data = [NSData dataWithBytes:bytes length:sizeof(bytes)];
    
    [self.peripheral writeValue:data forCharacteristic:self.carCharacteristic type:CBCharacteristicWriteWithoutResponse];
}

@end
