package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"io/ioutil"
	"strconv"
	"syscall"
)

// Stop...
func Stop(cname string) {
	pid, err := getPidByName(cname)
	if err != nil {
		logrus.Errorf("get container pid by name %s error %v", cname, err)
		return
	}
	pint, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("convert pid from string to int error %v", err)
		return
	}

	if err := syscall.Kill(pint, syscall.SIGTERM); err != nil {
		logrus.Errorf("stop container %s error %v", cname, err)
		return
	}

	info, err := getInfoByName(cname)
	if err != nil {
		logrus.Errorf("get container %s info error %v", cname, err)
		return
	}

	info.Status = common.Stop
	info.Pid = " "
	cbytes, err := json.Marshal(info)
	if err != nil {
		logrus.Errorf("json marshal %s error %v", cname, err)
		return
	}
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	cfp := dir + common.ContainerInfoFileName
	if err := ioutil.WriteFile(cfp, cbytes, 0622); err != nil {
		logrus.Errorf("write file %s error %v", cfp, err)
	}

}
