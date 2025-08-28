package repository

type AdminRepositoryInterface interface {
	GetMetrics() error
	GetLogs() error
	GetSystemInfo() error
}
