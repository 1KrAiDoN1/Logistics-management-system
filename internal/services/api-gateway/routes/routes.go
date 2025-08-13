package routes

import (
	"logistics/internal/services/api-gateway/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, h *handler.Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.SignUp)
	}
}
