package network

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go-docker/common"
	"net"
	"os"
	"path"
	"strings"
)

// IPAM...
type IPAM struct {
	// allocator path
	SubnetAllocatorPath string
	// key: net, value: bitmap
	Subnets *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: common.DefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	conf, err := os.Open(ipam.SubnetAllocatorPath)
	defer conf.Close()
	if err != nil {
		return err
	}
	sjson := make([]byte, 2000)
	n, err := conf.Read(sjson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(sjson[:n], ipam.Subnets)
	if err != nil {
		log.Errorf("error dump allocation info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	dir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dir, 0644)
		} else {
			return err
		}
	}

	conf, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer conf.Close()
	if err != nil {
		return err
	}
	ijson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = conf.Write(ijson)
	if err != nil {
		return err
	}
	return err
}

// Allocate...
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ipam.Subnets = &map[string]string{}

	err = ipam.load()
	if err != nil {
		log.Errorf("error load allocation info, %v", err)
	}
	// 127.0.0.0/8 mask is 255.0.0.0
	one, size := subnet.Mask.Size()
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// unused ip address
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP
		}

		for t := uint(4); t > 0; t -= 1 {
			[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
		}

		ip[3] += 1
		break
	}
	ipam.dump()
	return
}

// Release ...
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		log.Errorf("error dump allocation inof, %v", err)
	}

	c := 0
	ip := ipaddr.To4()
	ip[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(ip[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.dump()
	return nil
}
