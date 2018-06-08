package service

// SupportService is the interface that a support service for the environment must implement
type SupportService interface {
	Init() error
	Download() error
	Healthy() (bool, error)
	Read() (bool, error)
	Write() error
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
