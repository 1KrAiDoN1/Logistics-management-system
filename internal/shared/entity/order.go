package entity

// Order - основная структура заказа
type Order struct {
	ID              int64       `json:"id" db:"id"`
	UserID          int64       `json:"user_id" db:"user_id"`
	Status          OrderStatus `json:"status" db:"status"`
	DeliveryAddress string      `json:"delivery_address" db:"delivery_address"`
	Items           []GoodsItem `json:"items"`
	TotalAmount     float64     `json:"total_amount" db:"total_amount"`
	DriverID        *int64      `json:"driver_id,omitempty" db:"driver_id"`
	CreatedAt       int64       `json:"created_at" db:"created_at"`
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
type GoodsItem struct {
	ProductID   int64   `json:"product_id,omitempty" db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	Price       float64 `json:"price" db:"price"`
	Quantity    int32   `json:"quantity" db:"quantity"`
	TotalPrice  float64 `json:"total_price" db:"total_price"`
	LastUpdated int64   `json:"last_updated,,omitempty" db:"last_updated"`
}
