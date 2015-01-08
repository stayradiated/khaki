package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "khaki"
	app.Usage = "start the BLE peripheral"
	app.Version = "0.1.0"
	app.Author = "George Czabania"
	app.Email = "george@czabania.com"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "secret",
			Value: "hunter2",
			Usage: "secret key used to authenticate clients",
		},
		cli.BoolFlag{
			Name:  "public",
			Usage: "disable authentication (useful when debugging)",
		},
	}

	app.Action = func(c *cli.Context) {
		StartPeripheral(&PeripheralConfig{
			Secret: c.String("secret"),
			Public: c.Bool("public"),
		})
	}

	app.Run(os.Args)
}
