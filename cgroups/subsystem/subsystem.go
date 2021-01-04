package subsystem

// ResourceConfig ...
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

// Subsystem ...
type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	Subsystems = []Subsystem{
		&MemorySubSystem{},
		&CpuSubSystem{},
		&CpusetSubSystem{},
	}
)
