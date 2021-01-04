package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"io/ioutil"
	"os"
)

// Logs...
func Logs(cname string) {
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	lfile := dir + common.ContainerLogFileName
	// open log file
	file, err := os.Open(lfile)
	defer file.Close()
	if err != nil {
		logrus.Errorf("log container open file %s error %v", lfile, err)
		return
	}

	// read content
	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("log container read file %s error %v", lfile, err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}
