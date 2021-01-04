package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"os/exec"
)

// CommitContainer ...
func Commit(cname string, iname string) {
	mntURL := common.MntPath + cname + "/"
	tar := common.RootPath + iname + ".tar"
	fmt.Printf("%s", tar)
	if _, err := exec.Command("tar", "-czf", tar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s  error %v", mntURL, err)
	}
}
