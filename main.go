package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = `go-docker is a simple container runtime implementation.`

func main() {
	app := cli.NewApp()
	app.Name = "go-docker"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
	}

	app.Before = func(ctx *cli.Context) error {
		// log as json instead of the default ascii formatter
		log.SetFormatter(&log.JSONFormatter{})

		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
