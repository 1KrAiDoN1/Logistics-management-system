package app

import (
	"fmt"
	"log/slog"
	orderservice "logistics/internal/services/order-service"
	"logistics/pkg/lib/utils"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type OrderGRPCApp struct {
	log             *slog.Logger
	gRPCServer      *grpc.Server
	OrderGRPCConfig utils.ServiceConfig
}

func NewApp(log *slog.Logger, orderGRPCService *orderservice.OrderGRPCService, orderGRPCConfig utils.ServiceConfig) *OrderGRPCApp {
	gRPCServer := grpc.NewServer()
	orderservice.RegisterOrderServiceServer(gRPCServer, orderGRPCService)
	reflection.Register(gRPCServer)

	return &OrderGRPCApp{
		log:             log,
		gRPCServer:      gRPCServer,
		OrderGRPCConfig: orderGRPCConfig,
	}
}

func (a *OrderGRPCApp) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", a.OrderGRPCConfig.Address))
	if err != nil {
		return fmt.Errorf("failed to starting OrderGRPCServer: %w", err)
	}

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		a.log.Info("Order gRPC server is running on port: " + a.OrderGRPCConfig.Address)
		if err := a.gRPCServer.Serve(l); err != nil {
			serverErr <- fmt.Errorf("order gRPC server error: %w", err)
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
		a.log.Info("Shutting down...", "Received signal: ", sig)
		a.gRPCServer.GracefulStop()
		a.log.Info("Order gRPC server stopped")

	}
	return nil

}
