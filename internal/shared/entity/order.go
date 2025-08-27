package entity

import "time"

// Order - основная структура заказа
type Order struct {
	ID              int64       `json:"id" db:"id"`
	UserID          int64       `json:"user_id" db:"user_id"`
	Status          OrderStatus `json:"status" db:"status"`
	DeliveryAddress string      `json:"delivery_address" db:"delivery_address"`
	Items           []OrderItem `json:"items"`
	TotalAmount     float64     `json:"total_amount" db:"total_amount"`
	DriverID        *int64      `json:"driver_id" db:"driver_id"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
}

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"     // создан, ждет проверки склада
	StatusConfirmed  OrderStatus = "confirmed"   // товар зарезервирован
	StatusRouteReady OrderStatus = "route_ready" // маршрут построен
	StatusAssigned   OrderStatus = "assigned"    // водитель назначен
	StatusInProgress OrderStatus = "in_progress" // в пути
	StatusDelivered  OrderStatus = "delivered"   // доставлен
	StatusCancelled  OrderStatus = "cancelled"   // отменен
	StatusFailed     OrderStatus = "failed"      // ошибка
)

// OrderItem - товар в заказе
type OrderItem struct {
	ID          int64   `json:"id" db:"id"`
	OrderID     int64   `json:"order_id" db:"order_id"`
	ProductID   int64   `json:"product_id" db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	Price       float64 `json:"price" db:"price"`
}
