package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Exec...
func Exec(cname string, commands []string) {
	pid, err := getPidByName(cname)
	if err != nil {
		logrus.Errorf("exec container getPidByName %s error %v", cname, err)
		return
	}

	cmdstr := strings.Join(commands, " ")
	logrus.Infof("container pid %s", pid)
	logrus.Infof("command %s", cmdstr)

	// important
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(common.EnvExecPid, pid)
	os.Setenv(common.EnvExecCmd, cmdstr)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("exec container %s error %v", cname, err)
	}
}

func getPidByName(cname string) (string, error) {
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	cpath := dir + common.ContainerInfoFileName

	cbytes, err := ioutil.ReadFile(cpath)
	if err != nil {
		return "", err
	}
	var info ContainerInfo
	if err := json.Unmarshal(cbytes, &info); err != nil {
		return "", err
	}
	return info.Pid, nil
}
