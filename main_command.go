package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"go-docker/cgroups/subsystem"
	"go-docker/container"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `create a container with namespace and cgroups limit go-docker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	},

	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		res := &subsystem.ResourceConfig{
			MemoryLimit: ctx.String("m"),
			CpuSet:      ctx.String("cpuset"),
			CpuShare:    ctx.String("cpushare"),
		}
		var commands []string
		for _, arg := range ctx.Args().Tail() {
			commands = append(commands, arg)
		}
		tty := ctx.Bool("ti")
		Run(tty, commands, res)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "init container process run user's process in container",

	Action: func(ctx *cli.Context) error {
		log.Infof("init come on")
		cmd := ctx.Args().Get(0)
		log.Infof("command %s", cmd)
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}
