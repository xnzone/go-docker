package subsystem

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// GetCgroupPath ...
func GetCgroupPath(name string, cpath string, autoCreate bool) (string, error) {
	croot := FindCgroupMountpoint(name)
	if _, err := os.Stat(path.Join(croot, cpath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(croot, cpath), 0755); err == nil {

			} else {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(croot, cpath), nil
	} else {
		return "", fmt.Errorf("cgroup path err %v", err)
	}
}

// FindCgroupMountpoint ...
func FindCgroupMountpoint(name string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scans := bufio.NewScanner(f)
	for scans.Scan() {
		txt := scans.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == name {
				return fields[4]
			}
		}
	}
	if err := scans.Err(); err != nil {
		return ""
	}
	return ""
}
