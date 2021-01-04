package container

import (
	"github.com/sirupsen/logrus"
	"go-docker/common"
	"os"
	"os/exec"
	"strings"
)

// NewWorkSpace ...
func NewWorkSpace(volume string, cname string, iname string) {
	CreateReadOnlyLayer(iname)
	CreateWriteLayer(cname)
	CreateMountPoint(cname, iname)

	if volume != "" {
		volumes := volumeUrlExtract(volume)
		length := len(volumes)
		if length == 2 && volumes[0] != "" && volumes[1] != "" {
			MountVolume(volumes, cname)
			logrus.Infof("%q", volumes)
		} else {
			logrus.Infof("volume parameter input is not correct.")
		}
	}
}

// CreateReadOnlyLayer ...
func CreateReadOnlyLayer(iname string) error {
	tarurl := common.RootPath + iname + "/"
	imgurl := common.RootPath + iname + ".tar"
	exist, err := PathExists(tarurl)
	if err != nil {
		logrus.Infof("fail to judge whether dir %s exists. %v", tarurl, err)
		return err
	}

	if !exist {
		if err := os.MkdirAll(tarurl, 0622); err != nil {
			logrus.Errorf("mkdir %s error %v", tarurl, err)
			return err
		}

		if _, err := exec.Command("tar", "-xvf", imgurl, "-C", tarurl).CombinedOutput(); err != nil {
			logrus.Errorf("untar dir %s error %v", tarurl, err)
			return err
		}
	}
	return nil
}

// CreateWriteLayer create a container layer as an only writeable layer
func CreateWriteLayer(cname string) {
	writeURL := common.RootPath + common.WriteLayer + "/" + cname
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

// CreateMountPoint ...
func CreateMountPoint(cname string, iname string) error {
	// create mnt dir
	mnturl := common.MntPath + cname
	if err := os.Mkdir(mnturl, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", mnturl, err)
		return err
	}

	tmpwl := common.RootPath + common.WriteLayer + "/" + cname
	tmpil := common.RootPath + iname
	dirs := "dirs=" + tmpwl + ":" + tmpil
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mnturl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
		return err
	}
	return nil
}

// DeleteWorkSpace ...
func DeleteWorkSpace(volume, cname string) {
	if volume != "" {
		volumes := volumeUrlExtract(volume)
		length := len(volumes)
		if length == 2 && volumes[0] != "" && volumes[1] != "" {
			DeleteMountPointWithVolume(volumes, cname)
		} else {
			DeleteMountPoint(cname)
		}
	} else {
		DeleteMountPoint(cname)
	}
	DeleteWriteLayer(cname)
}

// DeleteMountPoint ...
func DeleteMountPoint(cname string) error {
	mntURL := common.MntPath + cname
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %s error %v", mntURL, err)
		return err
	}
	return nil
}

// DeleteWriteLayer ...
func DeleteWriteLayer(cname string) {
	writeURL := common.RootPath + common.WriteLayer + "/" + cname
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("remove dir %s error %v", writeURL, err)
	}
}

// DeleteMountPointWithVolume ...
func DeleteMountPointWithVolume(volumes []string, cname string) error {
	// umount volume
	mntURL := common.MntPath + cname
	curl := mntURL + "/" + volumes[1]
	if _, err := exec.Command("umount", curl).CombinedOutput(); err != nil {
		logrus.Errorf("umount volume %s failed. %v", curl, err)
		return err
	}
	// umount mountpoint
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("umount mountpoint failed. %v", err)
		return err
	}
	// delete mountpoint
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("remove mountpoint dir %s error. %v", mntURL, err)
	}
	return nil
}

// PathExists ...
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MountVolume ...
func MountVolume(volumes []string, cname string) error {
	parentURL := volumes[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Infof("mkdir parent dir %s error. %v", parentURL, err)
	}

	curl := volumes[1]
	mnturl := common.MntPath + cname
	cvurl := mnturl + "/" + curl
	if err := os.Mkdir(cvurl, 0777); err != nil {
		logrus.Infof("mkdir container dir %s error. %v", cvurl, err)
	}

	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", cvurl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("mount volume failed. %v", err)
		return err
	}
	return nil
}

// volumeUrlExtract extract volume url
func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}
