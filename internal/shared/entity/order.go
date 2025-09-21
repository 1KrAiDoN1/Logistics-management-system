package entity

// Order - основная структура заказа
// @Description Информация о заказе
type Order struct {
	ID              int64       `json:"id" db:"id" example:"1"`
	UserID          int64       `json:"user_id" db:"user_id" example:"123"`
	Status          OrderStatus `json:"status" db:"status" example:"pending"`
	DeliveryAddress string      `json:"delivery_address" db:"delivery_address" example:"ул. Пушкина, д. 10"`
	Items           []GoodsItem `json:"items"`
	TotalAmount     float64     `json:"total_amount" db:"total_amount" example:"15000.50"`
	DriverID        *int64      `json:"driver_id,omitempty" db:"driver_id" example:"456"`
	CreatedAt       int64       `json:"created_at" db:"created_at" example:"1694966400"`
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

// GoodsItem - товар в заказе
// @Description Товар в составе заказа
type GoodsItem struct {
	ProductID   int64   `json:"product_id,omitempty" db:"product_id" example:"789"`
	ProductName string  `json:"product_name" db:"product_name" example:"Ноутбук"`
	Price       float64 `json:"price" db:"price" example:"15000.00"`
	Quantity    int32   `json:"quantity" db:"quantity" example:"1"`
	TotalPrice  float64 `json:"total_price" db:"total_price" example:"15000.00"`
	LastUpdated int64   `json:"last_updated,,omitempty" db:"last_updated" example:"1694966400"`
}
