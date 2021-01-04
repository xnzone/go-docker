package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess ...
func NewParentProcess(tty bool, volume string, cname string, iname string) (*exec.Cmd, *os.File) {
	rpipe, wpipe, err := os.Pipe()
	if err != nil {
		logrus.Errorf("new pipe error %v", err)
		return nil, nil
	}
	// 调用自身，传入init参数，执行initCommand
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// print to log file
		dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
		if err := os.MkdirAll(dir, 0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error %v", dir, err)
			return nil, nil
		}
		lpath := dir + common.ContainerLogFileName
		lfile, err := os.Create(lpath)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error %v", lpath, err)
			return nil, nil
		}
		cmd.Stdout = lfile
	}
	cmd.ExtraFiles = []*os.File{rpipe}
	// replace /root/busybox as /root/mnt
	// mntURL := common.MntPath
	//rootURL := common.RootPath
	NewWorkSpace(volume, cname, iname)
	cmd.Dir = common.MntPath + cname

	return cmd, wpipe
}
