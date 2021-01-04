package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"os/exec"
)

// CommitContainer ...
func CommitContainer(name string) {
	mntURL := common.MntPath
	tar := common.RootPath + name + ".tar"
	fmt.Printf("%s", tar)
	if _, err := exec.Command("tar", "-czf", tar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s  error %v", mntURL, err)
	}
}
