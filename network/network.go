package network

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go-docker/common"
	"go-docker/container"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

type Network struct {
	Name    string     //name
	IpRange *net.IPNet // address
	Driver  string     // driver
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string
	Network     *Network
}

type NetDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

var (
	drivers  = map[string]NetDriver{}
	networks = map[string]*Network{}
)

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	gatewayIp, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIp

	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	return nw.dump(common.DefaultAllocatorPath)
}

func (nw *Network) dump(dpath string) error {
	if _, err := os.Stat(dpath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dpath, 0644)
		} else {
			return err
		}
	}

	nwpath := path.Join(dpath, nw.Name)
	nwfile, err := os.OpenFile(nwpath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf("error: %v", err)
		return err
	}
	defer nwfile.Close()

	nwjson, err := json.Marshal(nw)
	if err != nil {
		log.Errorf("error: %v", err)
		return err
	}
	_, err = nwfile.Write(nwjson)
	if err != nil {
		log.Errorf("error: %v", err)
		return err
	}
	return nil
}

func (nw *Network) load(dpath string) error {
	nwconf, err := os.Open(dpath)
	defer nwconf.Close()
	if err != nil {
		return err
	}

	nwjson := make([]byte, 200)
	n, err := nwconf.Read(nwjson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwjson[:n], nw)
	if err != nil {
		log.Errorf("error load nw info: %v", err)
		return err
	}
	return nil
}

func Connect(name string, cinfo *container.Info) error {
	nw, ok := networks[name]
	if !ok {
		return fmt.Errorf("no such network: %s", name)
	}
	ip, err := ipAllocator.Allocate(nw.IpRange)
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, name),
		IPAddress:   ip,
		Network:     nw,
		PortMapping: cinfo.PortMapping,
	}

	if err = drivers[nw.Driver].Connect(nw, ep); err != nil {
		return err
	}

	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	return configPortMapping(ep, cinfo)
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")

	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("flush error %v", err)
	}
}

func DeleteNetwork(name string) error {
	nw, ok := networks[name]
	if !ok {
		return fmt.Errorf("no such network: %s", name)
	}

	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("error remove network gateway ip: %s", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error remove network driver error: %s", err)
	}
	return nw.remove(common.DefaultNetworkPath)
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.Info) error {
	plink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	defer enterContainerNetns(&plink, cinfo)()
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v, %s", ep.Network, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	defaultRoute := &netlink.Route{
		LinkIndex: plink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}
	return nil
}

func configPortMapping(ep *Endpoint, cinfo *container.Info) error {
	for _, pm := range ep.PortMapping {
		port := strings.Split(pm, ":")
		if len(port) != 2 {
			log.Errorf("port mapping format error, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			port[0], ep.IPAddress.String(), port[2])

		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables output, %v", output)
			continue
		}
	}
	return nil
}

func enterContainerNetns(enlink *netlink.Link, cinfo *container.Info) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace, %v", err)
	}

	nsfd := f.Fd()

	runtime.LockOSThread()
	if err = netlink.LinkSetNsFd(*enlink, int(nsfd)); err != nil {
		log.Errorf("error set link netns, %v", err)
	}

	origns, err := netns.Get()
	if err != nil {
		log.Errorf("error get current netns, %v", err)
	}
	if err = netns.Set(netns.NsHandle(nsfd)); err != nil {
		log.Errorf("error set netns, %v", err)
	}
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func Init() error {
	var bdriver = BridgeNetworkDriver{}
	drivers[bdriver.Name()] = &bdriver
	if _, err := os.Stat(common.DefaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(common.DefaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	filepath.Walk(common.DefaultNetworkPath, func(nwpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		_, name := path.Split(nwpath)
		nw := &Network{
			Name: name,
		}
		if err = nw.load(nwpath); err != nil {
			log.Errorf("error load network: %s", err)
		}
		networks[name] = nw
		return nil
	})

	return nil
}

func (nw *Network) remove(dpath string) error {
	if _, err := os.Stat(path.Join(dpath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dpath, nw.Name))
	}
}
