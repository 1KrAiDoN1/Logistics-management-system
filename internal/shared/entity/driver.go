package entity

// Driver - структура водителя
type Driver struct {
	ID int64 `json:"id" db:"id"`
	// UserID        int64        `json:"user_id" db:"user_id"` // связь с пользователем
	Name          string       `json:"name" db:"name"`
	Phone         string       `json:"phone" db:"phone"`
	Email         string       `json:"email" db:"email"`
	LicenseNumber string       `json:"license_number" db:"license_number"`
	Status        DriverStatus `json:"status" db:"status"`
}

type DriverStatus string

const (
	DriverStatusOffline     DriverStatus = "offline"
	DriverStatusAvailable   DriverStatus = "available"   // свободен, может взять заказ
	DriverStatusBusy        DriverStatus = "busy"        // выполняет заказ
	DriverStatusBreak       DriverStatus = "break"       // на перерыве
	DriverStatusUnavailable DriverStatus = "unavailable" // недоступен (болеет, отпуск)
)
