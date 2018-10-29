package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "flowdump"
	app.Usage = "TCP/UDP flow detail dumper"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "interface, i",
			Usage: "Monitoring network interface",
		},
	}

	app.Action = func(c *cli.Context) error {
		opts := newFlowDumpOpts()
		opts.deviceName = c.String("i")

		err := loop(opts)
		if err != nil {
			log.WithField("error", err).Error("Error in main loop")
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
