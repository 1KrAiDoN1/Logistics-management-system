package dto

import (
	"logistics/internal/shared/entity"
	"time"
)

// CreateOrderRequest - запрос на создание заказа
// @Description Запрос на создание нового заказа
type CreateOrderRequest struct {
	UserID          int64             `json:"user_id" validate:"required" example:"123"`
	DeliveryAddress string            `json:"delivery_address" validate:"required" example:"ул. Пушкина, д. 10"`
	Items           []CreateOrderItem `json:"items" validate:"required,min=1"`
}

type CreateOrderItem struct {
	ProductName string `json:"product_name" validate:"required" example:"Ноутбук"`
	Quantity    int32  `json:"quantity" validate:"required,min=1" example:"1"`
}

// CreateOrderResponse - ответ на создание заказа
// @Description Ответ после успешного создания заказа
type CreateOrderResponse struct {
	Order   *entity.Order `json:"order"`
	Message string        `json:"message" example:"Order created successfully"`
}

// OrderStatusResponse - ответ со статусом заказа
// @Description Информация о статусе заказа
type OrderStatusResponse struct {
	OrderID   int64              `json:"order_id" example:"456"`
	Status    entity.OrderStatus `json:"status" example:"in_progress"`
	Driver    *DriverInfo        `json:"driver,omitempty"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type DriverInfo struct {
	Name    string      `json:"name" example:"Иван Иванов"`
	Phone   string      `json:"phone" example:"+79123456789"`
	Rating  float64     `json:"rating" example:"4.8"`
	Vehicle VehicleInfo `json:"vehicle,omitempty"`
}

type VehicleInfo struct {
	Model        string `json:"model" example:"Toyota Camry"`
	LicensePlate string `json:"license_plate" example:"А123БВ77"`
	Type         string `json:"type" example:"седан"`
}
