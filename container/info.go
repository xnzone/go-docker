package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// Info
type Info struct {
	Pid         string `json:"pid"`        // init pid
	Id          string `json:"id"`         // container id
	Name        string `json:"name"`       // container name
	Command     string `json:"command"`    // init command
	CreatedTime string `json:"createTime"` // create time
	Status      string `json:"status"`     // container status
	Volume      string `json:"volume"`     // container volume
}

// RecordContainerInfo ...
func RecordContainerInfo(pid int, commands []string, cname string, volume string) error {
	info := &Info{
		Id:          GenContainerID(10),
		Pid:         strconv.Itoa(pid),
		Command:     strings.Join(commands, ""),
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:      common.Running,
		Name:        cname,
		Volume:      volume,
	}
	bt, err := json.Marshal(info)
	if err != nil {
		logrus.Errorf("record container info error %v", err)
		return err
	}
	str := string(bt)

	// gen container path info
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logrus.Errorf("mkdir error %s error %v", dir, err)
		return err
	}
	fileName := fmt.Sprintf("%s/%s", dir, common.ContainerInfoFileName)
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		logrus.Errorf("create file %s error %v", fileName, err)
		return err
	}
	if _, err = file.WriteString(str); err != nil {
		logrus.Errorf("file write string error %v", err)
		return err
	}
	return nil

}

// DeleteContainerInfo...
func DeleteContainerInfo(cname string) {
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	if err := os.RemoveAll(dir); err != nil {
		logrus.Errorf("remove dir %s error %v", dir, err)
	}
}

// ListContainers...
func ListContainers() {
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, "")
	dir = dir[:len(dir)-1]
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logrus.Errorf("read dir %s error %v", dir, err)
		return
	}

	var containers []*Info
	for _, file := range files {
		tmp, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("get container info error %v", err)
			continue
		}
		containers = append(containers, tmp)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tName\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("flush error %v", err)
		return
	}
}

// GenContainerID...
func GenContainerID(n int) string {
	lbytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = lbytes[rand.Intn(len(lbytes))]
	}
	return string(b)
}

func getContainerInfo(file os.FileInfo) (*Info, error) {
	cname := file.Name()
	cfdir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	cfdir = cfdir + common.ContainerInfoFileName
	// read config.json
	content, err := ioutil.ReadFile(cfdir)
	if err != nil {
		logrus.Errorf("read file %s error %v", cfdir, err)
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(content, &info); err != nil {
		logrus.Errorf("json unmarshal error %v", err)
		return nil, err
	}
	return &info, nil
}

func getInfoByName(cname string) (*Info, error) {
	dir := fmt.Sprintf(common.DefaultContainerInfoPath, cname)
	cfp := dir + common.ContainerInfoFileName
	cbytes, err := ioutil.ReadFile(cfp)
	if err != nil {
		logrus.Errorf("read file %s error %v", cfp, err)
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(cbytes, &info); err != nil {
		logrus.Errorf("GetContainerInfoByName unmarshal error %v", err)
		return nil, err
	}
	return &info, nil
}
