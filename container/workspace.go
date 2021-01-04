package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

// NewWorkSpace ...
func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)

	if volume != "" {
		volumes := volumeUrlExtract(volume)
		length := len(volumes)
		if length == 2 && volumes[0] != "" && volumes[1] != "" {
			MountVolume(rootURL, mntURL, volumes)
			logrus.Infof("%q", volumes)
		} else {
			logrus.Infof("volume parameter input is not correct.")
		}
	}
}

// CreateReadOnlyLayer ...
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		logrus.Infof("fail to judge where dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			logrus.Errorf("mkdir dir % s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			logrus.Errorf("utar dir %s error %v", busyboxTarURL, err)
		}
	}
}

// CreateWriteLayer create a container layer as an only writeable layer
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

// CreateMountPoint ...
func CreateMountPoint(rootURL string, mntURL string) {
	// create mnt dir
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", mntURL, err)
	}

	dirs := "dirs=" + rootURL + "writeLayer" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
}

// DeleteWorkSpace ...
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumes := volumeUrlExtract(volume)
		length := len(volumes)
		if length == 2 && volumes[0] != "" && volumes[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumes)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
	} else {
		DeleteMountPoint(rootURL, mntURL)
	}
	DeleteWriteLayer(rootURL)
}

// DeleteMountPoint ...
func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %s error %v", mntURL, err)
	}
}

// DeleteWriteLayer ...
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("remove dir %s error %v", writeURL, err)
	}
}

// DeleteMountPointWithVolume ...
func DeleteMountPointWithVolume(rootURL, mntURL string, volumes []string) {
	// umount volume
	curl := mntURL + volumes[1]
	cmd := exec.Command("umount", curl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount volume failed. %v", err)
	}
	// umount mountpoint
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount mountpoint failed. %v", err)
	}
	// delete mountpoint
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("remove mountpoint dir %s error. %v", mntURL, err)
	}

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
func MountVolume(rootURL string, mntURL string, volumes []string) {
	parentURL := volumes[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Infof("mkdir parent dir %s error. %v", parentURL, err)
	}

	curl := volumes[1]
	cvurl := mntURL + curl
	if err := os.Mkdir(cvurl, 0777); err != nil {
		logrus.Infof("mkdir container dir %s error. %v", cvurl, err)
	}

	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", cvurl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("mount volume failed. %v", err)
	}
}

// volumeUrlExtract extract volume url
func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}
