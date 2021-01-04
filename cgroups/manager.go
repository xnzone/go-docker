package cgroups

import (
	"github.com/sirupsen/logrus"
	"go-docker/cgroups/subsystem"
)

type CgroupManager struct {
	Path     string
	Resource *subsystem.ResourceConfig
}

// NewCgroupManager ...
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Apply ...
func (c *CgroupManager) Apply(pid int) error {
	for _, ins := range subsystem.Subsystems {
		ins.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystem.ResourceConfig) error {
	for _, ins := range subsystem.Subsystems {
		ins.Set(c.Path, res)
	}
	return nil
}

// Destroy ...
func (c *CgroupManager) Destroy() error {
	for _, ins := range subsystem.Subsystems {
		if err := ins.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}
