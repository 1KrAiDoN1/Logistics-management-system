package app

import (
	"fmt"
	"log/slog"
	auth_grpc_server "logistics/internal/services/auth-service/grpc"
	"logistics/pkg/lib/utils"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type AuthGRPCApp struct {
	log            *slog.Logger
	gRPCServer     *grpc.Server
	AuthGRPCConfig utils.ServiceConfig
}

func NewApp(log *slog.Logger, authGRPCService *auth_grpc_server.AuthGRPCServer, authGRPCConfig utils.ServiceConfig) *AuthGRPCApp {
	gRPCServer := grpc.NewServer()
	auth_grpc_server.RegisterAuthServiceServer(gRPCServer, authGRPCService)
	reflection.Register(gRPCServer)

	return &AuthGRPCApp{
		log:            &slog.Logger{},
		gRPCServer:     gRPCServer,
		AuthGRPCConfig: authGRPCConfig,
	}
}

func (a *AuthGRPCApp) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", a.AuthGRPCConfig.Address))
	if err != nil {
		return fmt.Errorf("failed to starting AuthGRPCServer: %w", err)
	}

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("gRPC server is running on port: " + a.AuthGRPCConfig.Address)
		if err := a.gRPCServer.Serve(l); err != nil {
			serverErr <- fmt.Errorf("gRPC server error: %w", err)
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
		slog.Info("Shutting down...", "Received signal: ", sig)
		a.gRPCServer.GracefulStop()
		slog.Info("gRPC server stopped")

	}
	return nil

}
