package app

import (
	"fmt"
	"log/slog"
	driverservice "logistics/internal/services/driver-service"
	"logistics/pkg/lib/utils"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type DriverGRPCApp struct {
	log              *slog.Logger
	gRPCServer       *grpc.Server
	DriverGRPCConfig utils.ServiceConfig
}

func NewApp(log *slog.Logger, driverGRPCService *driverservice.DriverGRPCService, driverGRPCConfig utils.ServiceConfig) *DriverGRPCApp {
	gRPCServer := grpc.NewServer()
	driverservice.RegisterDriverServiceServer(gRPCServer, driverGRPCService)
	reflection.Register(gRPCServer)

	return &DriverGRPCApp{
		log:              log,
		gRPCServer:       gRPCServer,
		DriverGRPCConfig: driverGRPCConfig,
	}
}

func (a *DriverGRPCApp) Run() error {
	l, err := net.Listen("tcp", a.DriverGRPCConfig.Address)
	if err != nil {
		return fmt.Errorf("failed to starting driverGRPCServer: %w", err)
	}

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		a.log.Info("driver gRPC server is running on port: " + a.DriverGRPCConfig.Address)
		if err := a.gRPCServer.Serve(l); err != nil {
			serverErr <- fmt.Errorf("driver gRPC server error: %w", err)
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
		a.log.Info("driver gRPC server stopped")

	}
	return nil

}
