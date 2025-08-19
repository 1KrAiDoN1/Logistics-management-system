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

func SetupUserRoutes(router *gin.RouterGroup, userHandler handler.UserHandlerInterface) {
	users := router.Group("/user")
	{
		users.GET("/profile", userHandler.GetUserProfile)
		users.DELETE("/account", userHandler.DeleteUser)
	}
}

func SetupOrderRoutes(router *gin.RouterGroup, orderHandler handler.OrderHandlerInterface) {
	orders := router.Group("/orders")
	{
		orders.POST("", orderHandler.CreateOrder)
		orders.GET("", orderHandler.GetOrders)
		orders.GET("/:order_id", orderHandler.GetOrderByID)
		orders.DELETE("/:order_id", orderHandler.CancelOrder)
		orders.GET("/:order_id/tracking", orderHandler.TrackOrder)
		orders.POST("/:order_id/assign-driver", orderHandler.AssignDriver)
	}
}

func SetupAdminRoutes(router *gin.RouterGroup, adminHandler handler.AdminHandlerInterface) {
	admin := router.Group("/admin")
	{
		admin.GET("/health", adminHandler.HealthCheck)
		admin.GET("/metrics", adminHandler.GetMetrics)
		admin.GET("/logs", adminHandler.GetLogs)
		admin.POST("/config/reload", adminHandler.ReloadConfig)
		admin.GET("/services/status", adminHandler.GetServicesStatus)
		admin.POST("/cache/clear", adminHandler.ClearCache)
		admin.GET("/system/info", adminHandler.GetSystemInfo)
	}
}

func SetupDeliveryRoutes(router *gin.RouterGroup, deliveryHandler handler.DeliveryHandlerInterface) {
	deliveries := router.Group("/deliveries")
	{
		deliveries.GET("", deliveryHandler.GetDeliveries)
		deliveries.GET("/:delivery_id", deliveryHandler.GetDeliveryByID)
		deliveries.PUT("/:delivery_id/status", deliveryHandler.UpdateDeliveryStatus)
		deliveries.GET("/:delivery_id/tracking", deliveryHandler.TrackDelivery)
		deliveries.POST("/:delivery_id/proof", deliveryHandler.UploadProofOfDelivery)
		deliveries.GET("/driver/:driver_id", deliveryHandler.GetDeliveriesByDriver)
		deliveries.GET("/route/:route_id", deliveryHandler.GetDeliveriesByRoute)
	}
}

func SetupWarehouseRoutes(router *gin.RouterGroup, warehouseHandler handler.WarehouseHandlerInterface) {
	warehouse := router.Group("/store")
	{
		warehouse.GET("/products", warehouseHandler.GetAvailableProducts) // Список товаров
	}
}
