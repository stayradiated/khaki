//
//  ViewController.h
//  Khaki
//
//  Created by George Czabania on 3/01/15.
//  Copyright (c) 2015 Mintcode. All rights reserved.
//

#import <UIKit/UIKit.h>
#import <CommonCrypto/CommonHMAC.h>

@import CoreBluetooth;

#define KHAKI_SERVICE_UUID [CBUUID UUIDWithString:@"54a64ddf-c756-4a1a-bf9d-14f2cac357ad"]
#define KHAKI_CAR_CHARACTERISTIC_UUID [CBUUID UUIDWithString:@"fd1c6fcc-3ca5-48a9-97e9-37f81f5bd9c5"]
#define KHAKI_AUTH_CHARACTERISTIC_UUID [CBUUID UUIDWithString:@"66e01614-13d1-40d6-a34f-c5360ba57698"]

@interface ViewController : UIViewController <CBCentralManagerDelegate, CBPeripheralDelegate>

@property (nonatomic, strong) CBCentralManager *centralManager;
@property (nonatomic, strong) CBPeripheral *peripheral;

@property (nonatomic, strong) CBCharacteristic *authCharacteristic;
@property (nonatomic, strong) CBCharacteristic *carCharacteristic;

@property (weak, nonatomic) IBOutlet UILabel *connectedLabel;

@end

