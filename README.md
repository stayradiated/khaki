# Khaki

> Unlock your car using your phone

## What You Need

- Raspberry Pi model B
- Bluetooth 4.0 USB adapter
- A spare key fob that can unlock your car
- An iPhone

## The Setup

We have a Raspbery Pi with Arch Linux for ARM installed on it.
On top of that we have the bluez libraries installed as well as the Go
compiler.

The Pi will be connected to the spare key fob, specifically the buttons that 
lock and unlock the car.

We will be using the github.com/paypal/gatt library to setup a BLE peripheral
that can be connected to from a phone.

## How It Will Work

As you approach the car, your iPhone will automatically connect to the car and
authenticate with it. Then as you approach a specified distance the iPhone will
send the command to unlock the car.

When you walk away from the car, the iPhone will send the command to lock the
car.

The iPhone app will run in the background so you can keep your phone in your
pocket and it will still work.

## Security

Uses a challenge-response system with a shared key encryption.

Start the peripheral server with a secret key, that only you know.

    ./khaki --secret=hunter2

When the phone connects, it immediately starts the authentication process. If
authentication isn't successful within 5 seconds, the Pi will close the
connection.

    S = Server/Central (Pi)
    C = Client/Peripheral (iPhone)

    // Client request challenge code
    C -> READ auth

    // Server generates random bytes for challenge
    S -> [3e 48 5a 8d fb 30 0d 54 71 6e a6 68 18 72 b0 34]

    // Client calculates HMAC of bytes, and sends it to to the server
    C -> WRITE auth [88 62 38 70 73 03 ea d9 92 d6 e4 96 29 03 a2 90 e6 f2 2c 9e 3d d8 90 9c f5 e6 c7 02 58 98 41 b9]

    // Server validates the HMAC and responds with status
    S -> [01]

    // Client can now lock/unlock the car
    C -> WRITE lock 01

    // Server responds with success
    S -> Success

The challenge code is hashed using HMAC with SHA256.

    h := hmac.New(sha256.New, secretKey)
    h.Write(challengeCode)
    result := h.Sum(nil)

This authentication process is only needed when the client first connects to
the server. From then on the connection is marked as authenticated, and the
client to unlock/lock the car.

When the client disconnects, the authentication status is reset. The server can
only have one connection at a time, so it shouldn't be possible for an attacker
to hijack an existing connection that has already been authenticated.

TODO: Research BLE security modes.

## Services & Characteristics

Very simple interface, with just one service with two characteristics.

- Car `54a64ddf-c756-4a1a-bf9d-14f2cac357ad`
    - Lock `WRITE` `fd1c6fcc-3ca5-48a9-97e9-37f81f5bd9c5`
    - Auth `READ,WRITE` `66e01614-13d1-40d6-a34f-c5360ba57698`
