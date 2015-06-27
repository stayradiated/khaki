package main

import (
	"os"

	"github.com/codegangsta/cli"
)

// main sets up the the CLI
func main() {
	app := cli.NewApp()

	// app options
	app.Name = "khaki"
	app.Usage = "start the BLE peripheral"
	app.Version = "0.1.0"
	app.Author = "George Czabania"
	app.Email = "george@czabania.com"

	// app flags
	app.Flags = []cli.Flag{

		// authentication secret
		cli.StringFlag{
			Name:  "secret",
			Value: "hunter2",
			Usage: "secret key used to authenticate clients",
		},

		// disable authentication
		cli.BoolFlag{
			Name:  "public",
			Usage: "disable authentication (useful when debugging)",
		},
	}

	// app action handler
	app.Action = func(c *cli.Context) {

		// Create a new peripheral
		peripheral := NewPeripheral(NewAuth(
			[]byte(c.String("secret")),
			c.Bool("public"),
		))

		// Start the peripheral
		peripheral.Start()
	}

	app.Run(os.Args)
}
