package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"os"
)

// Remove...
func Remove(cname string) {
	info, err := getInfoByName(cname)
	if err != nil {
		logrus.Errorf("get container %s info error %v", cname, err)
		return
	}

	// only stop container can be removed
	if info.Status != common.Stop {
		logrus.Errorf("couldn't remove running container")
		return
	}

	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	if err := os.RemoveAll(dir); err != nil {
		logrus.Errorf("remove file %s error %v", dir, err)
		return
	}
}
