package subsystem

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CpusetSubSystem ...
type CpusetSubSystem struct {
	apply bool
}

func (c *CpusetSubSystem) Name() string {
	return "cpuset"
}

func (c *CpusetSubSystem) Set(cpath string, res *ResourceConfig) error {
	scpath, err := GetCgroupPath(c.Name(), cpath, true)
	if err != nil {
		logrus.Errorf("get %s path, err: %v", cpath, err)
		return err
	}
	if res.CpuSet != "" {
		c.apply = true
		err := ioutil.WriteFile(path.Join(scpath, "cpuset.cpus"), []byte(res.CpuSet), 0644)
		if err != nil {
			logrus.Errorf("failed to write file cpuset.cpus, err: %+v", err)
			return err
		}
	}
	return nil
}

func (c *CpusetSubSystem) Remove(cpath string) error {
	scpath, err := GetCgroupPath(c.Name(), cpath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(scpath)
}

func (c *CpusetSubSystem) Apply(cpath string, pid int) error {
	if c.apply {
		scpath, err := GetCgroupPath(c.Name(), cpath, false)
		if err != nil {
			return err
		}
		tpath := path.Join(scpath, "tasks")
		err = ioutil.WriteFile(tpath, []byte(strconv.Itoa(pid)), os.ModePerm)
		if err != nil {
			logrus.Errorf("write pid to tasks, path: %s, pid: %d, err: %+v", tpath, pid, err)
			return err
		}
	}
	return nil
}
