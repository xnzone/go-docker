package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// RunContainerInitProcess ...
func RunContainerInitProcess() error {
	commands := readUserCommand()
	if commands == nil || len(commands) == 0 {
		return fmt.Errorf("run container get user command error, commands is nil")
	}

	// mount
	setUpMount()

	// modified. use exec.LookPath
	path, err := exec.LookPath(commands[0])
	if err != nil {
		logrus.Errorf("exec loop path error %v", err)
		return err
	}
	logrus.Infof("find path %s", path)
	if err := syscall.Exec(path, commands[0:], os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

// init mount point
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("get current location error %v", err)
		return
	}
	logrus.Infof("current location is %s", pwd)
	pivotRoot(pwd)

	// mount proc
	dmf := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(dmf), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")

}

// pivotRoot change container root path and make old system still
func pivotRoot(root string) error {
	// mount root again, bind mount means change a new mount point
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}
	// create rootfs/.pivot_root to save old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// new rootfs mount on pivot_root and old root mount on rootfs/.pivot_root
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// change current workspace to root path
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// unmount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	return os.Remove(pivotDir)

}
