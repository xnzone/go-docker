package network

import (
	"go-docker/container"
	"testing"
)

func TestBridgeInit(t *testing.T) {
	d := BridgeNetworkDriver{}
	_, err := d.Create("192.168.0.1/24", "test-bridge")
	t.Logf("err: %v", err)
}

func TestBridgeConnect(t *testing.T) {
	ep := Endpoint{
		ID: "test container",
	}

	n := Network{
		Name: "test-bridge",
	}

	d := BridgeNetworkDriver{}
	err := d.Connect(&n, &ep)
	t.Logf("err: %v", err)
}

func TestNetworkConnect(t *testing.T) {
	cinfo := &container.Info{
		Id:  "test-container",
		Pid: "48293",
	}

	d := BridgeNetworkDriver{}
	n, err := d.Create("192.168.0.1/24", "test-bridge")
	t.Logf("err: %v", err)
	Init()
	networks[n.Name] = n
	err = Connect(n.Name, cinfo)
	t.Logf("err: %v", err)
}

func TestLoad(t *testing.T) {
	n := Network{
		Name: "test-bridge",
	}
	n.load("/var/run/go-docker/network/testbridge")

	t.Logf("network: %v", n)
}
