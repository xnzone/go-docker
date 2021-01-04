package main

import (
	"github.com/sirupsen/logrus"
	"go-docker/cgroups"
	"go-docker/cgroups/subsystem"
	"go-docker/container"
	"os"
	"strings"
)

// Run ...
func Run(tty bool, commands []string, res *subsystem.ResourceConfig, volume string, cname string, iname string) {
	parent, wpipe := container.NewParentProcess(tty, volume, cname, iname)

	if parent == nil {
		logrus.Errorf("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}

	// record container info
	if err := container.RecordContainerInfo(parent.Process.Pid, commands, cname, volume); err != nil {
		logrus.Errorf("record container info error %v", err)
		return
	}

	cmanager := cgroups.NewCgroupManager("go-docker")
	defer cmanager.Destroy()

	cmanager.Set(res)
	cmanager.Apply(parent.Process.Pid)

	sendInitCommand(commands, wpipe)
	if tty {
		parent.Wait()
		container.DeleteContainerInfo(cname)
	}

	container.DeleteWorkSpace(volume, cname)
	os.Exit(-1)
}

func sendInitCommand(commands []string, wpipe *os.File) {
	command := strings.Join(commands, " ")
	logrus.Infof("command all is %s", command)
	_, _ = wpipe.WriteString(command)
	_ = wpipe.Close()
}
