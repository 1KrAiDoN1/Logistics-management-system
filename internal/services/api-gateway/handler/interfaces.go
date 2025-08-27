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
}

type OrderHandlerInterface interface {
	CreateOrder(c *gin.Context)
	GetOrders(c *gin.Context)
	GetOrderByID(c *gin.Context)
	AssignDriver(c *gin.Context)
}

type DriverHandlerInterface interface {
}
type DeliveryHandlerInterface interface {
	GetDeliveries(c *gin.Context)
}

type AdminHandlerInterface interface {
	GetMetrics(c *gin.Context)
	GetLogs(c *gin.Context)
	GetSystemInfo(c *gin.Context)
}

type WarehouseHandlerInterface interface {
	GetAvailableProducts(c *gin.Context)
}
