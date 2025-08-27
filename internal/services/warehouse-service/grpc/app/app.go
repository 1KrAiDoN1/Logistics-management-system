package app

import (
	"fmt"
	"log/slog"
	warehouseservice "logistics/internal/services/warehouse-service"
	"logistics/pkg/lib/utils"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type WarehouseGRPCApp struct {
	log                 *slog.Logger
	gRPCServer          *grpc.Server
	WarehouseGRPCConfig utils.ServiceConfig
}

func NewApp(log *slog.Logger, warehouseGRPCService *warehouseservice.WarehouseGRPCService, warehouseGRPCConfig utils.ServiceConfig) *WarehouseGRPCApp {
	gRPCServer := grpc.NewServer()
	warehouseservice.RegisterWarehouseServiceServer(gRPCServer, warehouseGRPCService)
	reflection.Register(gRPCServer)

	return &WarehouseGRPCApp{
		log:                 log,
		gRPCServer:          gRPCServer,
		WarehouseGRPCConfig: warehouseGRPCConfig,
	}
}

func (a *WarehouseGRPCApp) Run() error {
	l, err := net.Listen("tcp", a.WarehouseGRPCConfig.Address)
	if err != nil {
		return fmt.Errorf("failed to starting WarehouseGRPCServer: %w", err)
	}

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		a.log.Info("Warehouse gRPC server is running on port: " + a.WarehouseGRPCConfig.Address)
		if err := a.gRPCServer.Serve(l); err != nil {
			serverErr <- fmt.Errorf("warehouse gRPC server error: %w", err)
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
		a.log.Info("Warehouse gRPC server stopped")

	}
	return nil

}
