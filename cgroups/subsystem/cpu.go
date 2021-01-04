package subsystem

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CpuSubSystem ...
type CpuSubSystem struct {
	apply bool
}

func (c *CpuSubSystem) Name() string {
	return "cpu"
}

func (c *CpuSubSystem) Set(cpath string, res *ResourceConfig) error {
	scpath, err := GetCgroupPath(c.Name(), cpath, true)
	if err != nil {
		logrus.Errorf("get %s path: %v", cpath, err)
		return err
	}
	if res.CpuShare != "" {
		c.apply = true
		err = ioutil.WriteFile(path.Join(scpath, "cpu.share"), []byte(res.CpuShare), 0644)
		if err != nil {
			logrus.Errorf("failed to write file cpu.share, err: +%v", err)
			return err
		}
	}
	return nil
}

func (c *CpuSubSystem) Remove(cpath string) error {
	scpath, err := GetCgroupPath(c.Name(), cpath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(scpath)
}

func (c *CpuSubSystem) Apply(cpath string, pid int) error {
	if c.apply {
		scpath, err := GetCgroupPath(c.Name(), cpath, false)
		if err != nil {
			return err
		}
		tpath := path.Join(scpath, "tasks")
		err = ioutil.WriteFile(tpath, []byte(strconv.Itoa(pid)), os.ModePerm)
		if err != nil {
			logrus.Errorf("write pid to tasks, path: %s,pid: %d, err: %v", tpath, pid, err)
			return err
		}
	}
	return nil
}
