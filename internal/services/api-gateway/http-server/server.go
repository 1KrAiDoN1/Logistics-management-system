package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	orderpb "logistics/api/protobuf/order_service"
	"logistics/configs"
	"logistics/internal/services/api-gateway/handler"
	"logistics/internal/services/api-gateway/middleware"
	"logistics/internal/services/api-gateway/routes"
	"logistics/pkg/lib/logger/slogger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	// HTTP сервер
	router *gin.Engine

	authGRPCClient authpb.AuthServiceClient

	handlers *handler.Handlers // Хендлеры, которые используют gRPC-клиенты.

	microservices_config *configs.MicroservicesConfig

	logger *slog.Logger
}

func NewServer(logger *slog.Logger, microservices_config *configs.MicroservicesConfig) *Server {
	router := gin.Default()

	authGRPCConn, err := grpc.NewClient(microservices_config.AuthGRPCServiceConfig.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to create gRPC client for auth service", slogger.Err(err))
		return nil
	}
	driverGRPCConn, err := grpc.NewClient(microservices_config.DriverGRPCServiceConfig.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to create gRPC client for driver service", slogger.Err(err))
		return nil
	}
	orderGRPCConn, err := grpc.NewClient(microservices_config.DriverGRPCServiceConfig.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to create gRPC client for order service", slogger.Err(err))
		return nil
	}
	warehouseGRPCConn, err := grpc.NewClient(microservices_config.WarehouseGRPCServiceConfig.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to create gRPC client for warehouse service", slogger.Err(err))
		return nil
	}

	// Создание клиентов
	authGRPCClient := authpb.NewAuthServiceClient(authGRPCConn)
	orderGRPCClient := orderpb.NewOrderServiceClient(orderGRPCConn)
	driverGRPCClient := driverpb.NewDriverServiceClient(driverGRPCConn)
	warehouseGRPCClient := warehousepb.NewWarehouseServiceClient(warehouseGRPCConn)

	handlers := handler.NewHandlers(logger, authGRPCClient, orderGRPCClient, driverGRPCClient, warehouseGRPCClient)
	return &Server{
		router: router,
		// authGRPCClient:       authGRPCClient,
		handlers:             handlers,
		microservices_config: microservices_config,
		logger:               logger,
	}
}

func (s *Server) Run() error {

	s.setupRoutes()

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		s.logger.Info("Starting HTTP server", slog.String("address", s.microservices_config.ApiGatewayConfig.HTTPServer.Address))
		if err := s.router.Run(s.microservices_config.ApiGatewayConfig.HTTPServer.Address); err != nil {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
		close(serverErr)
	}()

	// Ожидаем сигналы завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем до получения сигнала или ошибки сервера
	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		s.logger.Info("Received shutdown signal", slog.String("signal", sig.String()))

		// Получаем доступ к внутреннему http.Server
		srv := &http.Server{
			Addr:    s.microservices_config.ApiGatewayConfig.HTTPServer.Address,
			Handler: s.router,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Error("Server forced to shutdown", slogger.Err(err))
		}
		s.logger.Info("Server gracefully shut down")
		return nil
	}

}

func (s *Server) setupRoutes() {
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := s.router.Group("/api/v1")

	// Public routes
	routes.SetupAuthRoutes(api, s.handlers.AuthHandlerInterface)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(s.authGRPCClient))
	{
		routes.SetupUserRoutes(protected, s.handlers.UserHandlerInterface)
		// routes.SetupCategoryRoutes(protected, s.handlers.CategoryHandlerInterface)
		// routes.SetupExpenseRoutes(protected, s.handlers.ExpenseHandlerInterface)
		// routes.SetupBudgetRoutes(protected, s.handlers.BudgetHandlerInterface)
	}
}
