package service

// SupportService is the interface that a support service for the environment must implement
type SupportService interface {
	Init(string) error
	Download() error
	Healthy() (bool, error)
	Read(string) (bool, error)
	Write(string) error
	WriteServiceConfig() error
	Start() error
	Stop() error
	Restart() error
	Workspace() string
	SetWorkspace(string)
	Name() string
	SetName(string)
	ServiceConfig() string
	SetServiceConfig(string)
	HealthyTimeout() int
	SetHealthyTimeout(int)
}
