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
func Run(tty bool, commands []string, res *subsystem.ResourceConfig, volume string) {
	parent, wpipe := container.NewParentProcess(tty, volume)

	if parent == nil {
		logrus.Errorf("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}
	cmanager := cgroups.NewCgroupManager("go-docker")
	defer cmanager.Destroy()

	cmanager.Set(res)
	cmanager.Apply(parent.Process.Pid)

	sendInitCommand(commands, wpipe)
	parent.Wait()

	mntURl := "/root/mnt"
	rootURL := "/root"
	container.DeleteWorkSpace(rootURL, mntURl, volume)
	os.Exit(-1)
}

func sendInitCommand(commands []string, wpipe *os.File) {
	command := strings.Join(commands, " ")
	logrus.Infof("command all is %s", command)
	_, _ = wpipe.WriteString(command)
	_ = wpipe.Close()
}
