package service

type AdminServiceInterface interface {
	GetMetrics() error
	GetLogs() error
	GetSystemInfo() error
}
