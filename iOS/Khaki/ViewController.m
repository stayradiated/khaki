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
    self.centralManager = [[CBCentralManager alloc] initWithDelegate:self
                                                               queue:nil
                                                             options:@{
                                                                       CBCentralManagerOptionRestoreIdentifierKey: @"KhakiCentral"
                                                                       }];
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
        
        // Start scanning for services
        [self.centralManager scanForPeripheralsWithServices:@[KHAKI_SERVICE_UUID] options:nil];
        
        if (self.peripheral) {
            NSLog(@"Already have a peripheral");
            if (self.peripheral.state == CBPeripheralStateConnected) {
                
                // Check if we have discovered the service
                NSUInteger serviceIdx = [self.peripheral.services indexOfObjectPassingTest:^BOOL(CBService *obj, NSUInteger idx, BOOL *stop) {
                    return [obj.UUID isEqual:KHAKI_SERVICE_UUID];
                }];
                
                if (serviceIdx == NSNotFound) {
                    // We haven't discovered all the services yet
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
                    [services addObject:KHAKI_CAR_CHARACTERISTIC_UUID];
                } else {
                    self.carCharacteristic = service.characteristics[carCharIdx];
                }
                
                if (authCharIdx == NSNotFound) {
                    [services addObject:KHAKI_AUTH_CHARACTERISTIC_UUID];
                } else {
                    self.authCharacteristic = service.characteristics[authCharIdx];
                }
                
                if ([services count] > 0) {
                    // We haven't discovered all the characteristic yet!
                    [self.peripheral discoverCharacteristics:services forService:service];
                    return;
                }
                
            }
        }
        
    } else {
        NSLog(@"CoreBluetooth BLE hardware is powered off");
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
    self.connectedLabel.text = @"Connected";
    
    [peripheral setDelegate:self];
    [peripheral discoverServices:@[KHAKI_SERVICE_UUID]];
}

// Called when a peripheral cannot be connected
- (void)centralManager:(CBCentralManager *)central didFailToConnectPeripheral:(CBPeripheral *)peripheral error:(NSError *)error {
    NSLog(@"Failed to connect to peripheral %@ (%@)", peripheral, error);
}

// Called when a bluetooth peripheral is disconnected
- (void)centralManager:(CBCentralManager *)central didDisconnectPeripheral:(CBPeripheral *)peripheral error:(NSError *)error {
    NSLog(@"Disconnected peripheral %@ (%@)", peripheral, error);
    
    self.connectedLabel.text = @"Not Connected";
    
    // Maybe try scanning again?
    [self.centralManager scanForPeripheralsWithServices:@[KHAKI_SERVICE_UUID] options:nil];
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
        if ([characteristic.UUID isEqual:KHAKI_AUTH_CHARACTERISTIC_UUID]) {
            [self authenticate:characteristic];
        }
    }
}

#pragma mark - Buttons

- (IBAction)tapUnlockButton:(id)sender {
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
    
    if (self.peripheral == nil || self.carCharacteristic == nil) {
        NSLog(@"Not yet connected...");
    }
    
    const unsigned char bytes[] = {2};
    NSData *data = [NSData dataWithBytes:bytes length:sizeof(bytes)];
    
    [self.peripheral writeValue:data forCharacteristic:self.carCharacteristic type:CBCharacteristicWriteWithoutResponse];
}

@end
