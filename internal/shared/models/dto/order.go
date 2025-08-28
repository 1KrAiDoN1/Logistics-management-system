package dto

import (
	"logistics/internal/shared/entity"
	"time"
)

// CreateOrderRequest - запрос на создание заказа
type CreateOrderRequest struct {
	UserID          int64             `json:"user_id" validate:"required"`
	DeliveryAddress string            `json:"delivery_address" validate:"required"`
	PickupAddress   string            `json:"pickup_address" validate:"required"`
	Items           []CreateOrderItem `json:"items" validate:"required,min=1"`
	Priority        string            `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
}

type CreateOrderItem struct {
	ProductID   int64   `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name" validate:"required"`
	Quantity    int32   `json:"quantity" validate:"required,min=1"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

// CreateOrderResponse - ответ на создание заказа
type CreateOrderResponse struct {
	Order   *entity.Order `json:"order"`
	Message string        `json:"message"`
}

// OrderStatusResponse - ответ со статусом заказа
type OrderStatusResponse struct {
	OrderID   int64              `json:"order_id"`
	Status    entity.OrderStatus `json:"status"`
	Driver    *DriverInfo        `json:"driver,omitempty"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type DriverInfo struct {
	ID      int64        `json:"id"`
	Name    string       `json:"name"`
	Phone   string       `json:"phone"`
	Rating  float64      `json:"rating"`
	Vehicle *VehicleInfo `json:"vehicle,omitempty"`
}

type VehicleInfo struct {
	Model        string `json:"model"`
	LicensePlate string `json:"license_plate"`
	Type         string `json:"type"`
}
