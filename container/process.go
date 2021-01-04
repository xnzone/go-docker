package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// RunContainerInitProcess ...
func RunContainerInitProcess() error {
	commands := readUserCommand()
	if commands == nil || len(commands) == 0 {
		return fmt.Errorf("run container get user command error, commands is nil")
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
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

// NewParentProcess ...
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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
	}
	cmd.ExtraFiles = []*os.File{rpipe}
	return cmd, wpipe
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
