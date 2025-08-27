package routes

import (
	"logistics/internal/services/api-gateway/handler"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup, authHandler handler.AuthHandlerInterface) {
	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", authHandler.SignUp)
		auth.POST("/sign-in", authHandler.SignIn)
		auth.POST("/logout", authHandler.Logout)
	}
}

func SetupOrderRoutes(router *gin.RouterGroup, orderHandler handler.OrderHandlerInterface) {
	orders := router.Group("/orders")
	{
		orders.POST("", orderHandler.CreateOrder)
		orders.GET("", orderHandler.GetOrders)
		orders.GET("/:order_id", orderHandler.GetOrderByID)
		orders.POST("/:order_id/assign-driver", orderHandler.AssignDriver)
		orders.GET("/delivery", orderHandler.GetDeliveries)
	}
}

func SetupAdminRoutes(router *gin.RouterGroup, adminHandler handler.AdminHandlerInterface) {
	admin := router.Group("/admin")
	{
		admin.GET("/metrics", adminHandler.GetMetrics)
		admin.GET("/logs", adminHandler.GetLogs)
		admin.GET("/system/info", adminHandler.GetSystemInfo)
	}
}

func SetupWarehouseRoutes(router *gin.RouterGroup, warehouseHandler handler.WarehouseHandlerInterface) {
	warehouse := router.Group("/store")
	{
		warehouse.GET("/products", warehouseHandler.GetAvailableProducts) // Список товаров
	}
}
