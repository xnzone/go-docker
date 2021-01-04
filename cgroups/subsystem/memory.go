package subsystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// MemorySubSystem ...
type MemorySubSystem struct {
}

func (s *MemorySubSystem) Set(cpath string, res *ResourceConfig) error {
	if scpath, err := GetCgroupPath(s.Name(), cpath, true); err == nil {
		if res.MemoryLimit != "" {
			if err := ioutil.WriteFile(path.Join(scpath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *MemorySubSystem) Remove(cpath string) error {
	if scpath, err := GetCgroupPath(s.Name(), cpath, false); err == nil {
		return os.Remove(scpath)
	} else {
		return err
	}
}

func (s *MemorySubSystem) Apply(cpath string, pid int) error {
	if scpath, err := GetCgroupPath(s.Name(), cpath, false); err != nil {
		if err := ioutil.WriteFile(path.Join(scpath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get croup %s error: %v", cpath, err)
	}
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}
