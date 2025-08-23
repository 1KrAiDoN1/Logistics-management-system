package entity

import "time"

// Driver - структура водителя
type Driver struct {
	ID              int64        `json:"id" db:"id"`
	UserID          int64        `json:"user_id" db:"user_id"` // связь с пользователем
	Name            string       `json:"name" db:"name"`
	Phone           string       `json:"phone" db:"phone"`
	Email           string       `json:"email" db:"email"`
	LicenseNumber   string       `json:"license_number" db:"license_number"`
	Status          DriverStatus `json:"status" db:"status"`
	CurrentLocation *Location    `json:"current_location"`
	VehicleID       *int64       `json:"vehicle_id" db:"vehicle_id"`
	MaxWeight       float64      `json:"max_weight" db:"max_weight"` // максимальная загрузка кг
	MaxVolume       float64      `json:"max_volume" db:"max_volume"` // максимальный объем м³
	WorkingHours    WorkingHours `json:"working_hours"`
	Rating          float64      `json:"rating" db:"rating"`
	CompletedOrders int32        `json:"completed_orders" db:"completed_orders"`
	ActiveOrderID   *int64       `json:"active_order_id" db:"active_order_id"`
	LastActiveAt    time.Time    `json:"last_active_at" db:"last_active_at"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

type DriverStatus string

const (
	DriverStatusOffline     DriverStatus = "offline"
	DriverStatusAvailable   DriverStatus = "available"   // свободен, может взять заказ
	DriverStatusBusy        DriverStatus = "busy"        // выполняет заказ
	DriverStatusBreak       DriverStatus = "break"       // на перерыве
	DriverStatusUnavailable DriverStatus = "unavailable" // недоступен (болеет, отпуск)
)

type Location struct {
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	Address   string    `json:"address" db:"address"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type WorkingHours struct {
	StartTime string `json:"start_time"` // "09:00"
	EndTime   string `json:"end_time"`   // "18:00"
	TimeZone  string `json:"timezone"`   // "UTC+3"
	WorkDays  []int  `json:"work_days"`  // [1,2,3,4,5] пн-пт
}

// Vehicle - транспортное средство
type Vehicle struct {
	ID           int64         `json:"id" db:"id"`
	DriverID     *int64        `json:"driver_id" db:"driver_id"`
	Model        string        `json:"model" db:"model"`
	LicensePlate string        `json:"license_plate" db:"license_plate"`
	MaxWeight    float64       `json:"max_weight" db:"max_weight"`
	MaxVolume    float64       `json:"max_volume" db:"max_volume"`
	Type         VehicleType   `json:"type" db:"type"`
	Status       VehicleStatus `json:"status" db:"status"`
}

type VehicleType string

const (
	VehicleTypeCar   VehicleType = "car"
	VehicleTypeTruck VehicleType = "truck"
	VehicleTypeVan   VehicleType = "van"
)

type VehicleStatus string

const (
	VehicleStatusActive      VehicleStatus = "active"
	VehicleStatusInactive    VehicleStatus = "inactive"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
)
