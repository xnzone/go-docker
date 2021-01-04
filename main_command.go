package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"go-docker/cgroups/subsystem"
	"go-docker/common"
	"go-docker/container"
	"go-docker/network"
	"os"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `create a container with namespace and cgroups limit go-docker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		// add volume tag
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		// add -d tag
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		// -name to name container
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		// env
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},
		// network
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
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
		detach := ctx.Bool("d")
		volume := ctx.String("v")
		// tty and detach can't exist at the same time
		if tty && detach {
			return fmt.Errorf("ti and d param can not both provided")
		}
		cname := ctx.String("name")
		// get image name
		iname := ctx.Args().Get(0)

		envs := ctx.StringSlice("e")

		net := ctx.String("net")
		ports := ctx.StringSlice("p")

		Run(tty, commands, res, volume, cname, iname, envs, net, ports)
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
		return container.RunContainerInitProcess()
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		cname := ctx.Args().Get(0)
		iname := ctx.Args().Get(1)
		// commit container
		container.Commit(cname, iname)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(ctx *cli.Context) error {
		container.ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("please input your container name")
		}
		cname := ctx.Args().Get(0)
		container.Logs(cname)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(ctx *cli.Context) error {
		if os.Getenv(common.EnvExecPid) != "" {
			log.Infof("pid callback pid %d", os.Getgid())
			return nil
		}

		if len(ctx.Args()) < 2 {
			return fmt.Errorf("missing container name or command")
		}
		cname := ctx.Args().Get(0)
		var commands []string
		for _, arg := range ctx.Args().Tail() {
			commands = append(commands, arg)
		}
		container.Exec(cname, commands)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		cname := ctx.Args().Get(0)
		container.Stop(cname)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("misssing container name")
		}
		cname := ctx.Args().Get(0)
		container.Remove(cname)
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(ctx *cli.Context) error {
				if len(ctx.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				err := network.Init()
				if err != nil {
					log.Errorf("network init failed err %v", err)
					return err
				}
				err = network.CreateNetwork(ctx.String("driver"), ctx.String("subnet"), ctx.Args().Get(0))
				if err != nil {
					return fmt.Errorf("create network err %v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container network",
			Action: func(ctx *cli.Context) error {
				err := network.Init()
				if err != nil {
					log.Errorf("network init failed err %v", err)
					return err
				}
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove container network",
			Action: func(ctx *cli.Context) error {
				if len(ctx.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}

				err := network.Init()
				if err != nil {
					log.Errorf("network init failed err %v", err)
					return err
				}

				err = network.DeleteNetwork(ctx.Args().Get(0))
				if err != nil {
					return fmt.Errorf("remove network err %v", err)
				}
				return nil
			},
		},
	},
}
