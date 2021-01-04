package common

const (
	RootPath   = "/root/"
	MntPath    = "/root/mnt/"
	WriteLayer = "writeLayer"
)

const (
	Running = "running"
	Stop    = "stopped"
	Exit    = "exited"
)

const (
	DefaultContainerInfoPath = "/var/run/go-docker/%s/"
	ContainerInfoFileName    = "config.json"
	ContainerLogFileName     = "container.log"
)

const (
	EnvExecPid = "docker_pid"
	EnvExecCmd = "docker_cmd"
)

const (
	DefaultAllocatorPath = "/var/run/go-docker/network/ipam/subnet.json"
	DefaultNetworkPath   = "/var/run/go-docker/network/network/"
)
