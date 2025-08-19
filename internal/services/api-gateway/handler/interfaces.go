package handler

import "github.com/gin-gonic/gin"

type AuthHandlerInterface interface {
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	Logout(c *gin.Context)
}

type UserHandlerInterface interface {
	GetUserProfile(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetUserStats(c *gin.Context)
}

type OrderHandlerInterface interface {
	CreateOrder(c *gin.Context)
	GetOrders(c *gin.Context)
	GetOrderByID(c *gin.Context)
	UpdateOrder(c *gin.Context)
	CancelOrder(c *gin.Context)
	UpdateOrderStatus(c *gin.Context)
	TrackOrder(c *gin.Context)
	AssignDriver(c *gin.Context)
	GetOrdersByCustomer(c *gin.Context)
	GetOrdersByDriver(c *gin.Context)
	GetOrderAnalytics(c *gin.Context)
}

type DeliveryHandlerInterface interface {
	GetDeliveries(c *gin.Context)
	GetDeliveryByID(c *gin.Context)
	GetDeliveryByOrderID(c *gin.Context)
	GetDeliveryByDriverID(c *gin.Context)
	UpdateDelivery(c *gin.Context)
	CancelDelivery(c *gin.Context)
	AssignDelivery(c *gin.Context)
	GetDeliveryAnalytics(c *gin.Context)
	GetDeliveryStatus(c *gin.Context)
}

type AdminHandlerInterface interface {
	HealthCheck(c *gin.Context)
	GetMetrics(c *gin.Context)
	GetLogs(c *gin.Context)
	ReloadConfig(c *gin.Context)
	GetServicesStatus(c *gin.Context)
	ClearCache(c *gin.Context)
	GetSystemInfo(c *gin.Context)
}

type WarehouseHandlerInterface interface {
	GetAvailableProducts(c *gin.Context)
}
