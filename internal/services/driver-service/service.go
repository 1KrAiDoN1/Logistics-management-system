package driverservice

import (
	"context"
	"encoding/json"
	"log/slog"
	driverpb "logistics/api/protobuf/driver_service"
	kfk "logistics/internal/kafka"
	"logistics/internal/services/driver-service/domain"
	"logistics/internal/shared/entity"
	"logistics/pkg/lib/logger/slogger"
	"math/rand"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DriverGRPCService struct {
	driverpb.UnimplementedDriverServiceServer
	driverRepo    domain.DriverRepositoryInterface
	logger        *slog.Logger
	kafkaProducer *kfk.KafkaProducer
}

func NewDriverGRPCService(logger *slog.Logger, driverRepo domain.DriverRepositoryInterface, kafkaProducer *kfk.KafkaProducer) *DriverGRPCService {
	return &DriverGRPCService{
		driverRepo:    driverRepo,
		logger:        logger,
		kafkaProducer: kafkaProducer,
	}
}

func RegisterDriverServiceServer(s *grpc.Server, srv *DriverGRPCService) {
	driverpb.RegisterDriverServiceServer(s, srv)
}

func (d *DriverGRPCService) FindSuitableDriver(ctx context.Context, req *driverpb.FindDriverRequest) (*driverpb.FindDriverResponse, error) {
	availableDriversResp, err := d.GetAvailableDrivers(ctx, &emptypb.Empty{})
	if err != nil {
		d.logger.Error("failed to get available drivers for finding suitable driver",
			slog.String("status", "error"), slogger.Err(err))
		return nil, status.Errorf(codes.Internal, "failed to find suitable driver: %v", err)
	}

	// Проверяем, есть ли доступные водители
	if len(availableDriversResp.Drivers) == 0 {
		d.logger.Warn("no available drivers found", slog.String("status", "warning"))
		return &driverpb.FindDriverResponse{
			Driver:  nil,
			Success: false,
			Message: "No available drivers found",
		}, nil
	}

	// Выбираем случайного водителя из списка
	rand.NewSource(time.Now().UnixNano())
	randomIndex := rand.Intn(len(availableDriversResp.Drivers))
	selectedDriver := availableDriversResp.Drivers[randomIndex]

	d.logger.Info("suitable driver found",
		slog.String("driver_id", strconv.Itoa(int(selectedDriver.DriverId))),
		slog.String("driver_name", selectedDriver.Name))

	msg := entity.DriverKafka{
		ID:            selectedDriver.DriverId,
		Name:          selectedDriver.Name,
		Phone:         selectedDriver.Phone,
		LicenseNumber: selectedDriver.Vehicle.LicensePlate,
		Car:           selectedDriver.Vehicle.Model,
	}

	messageBytes, err := json.Marshal(msg)
	if err != nil {
		d.logger.Error("Failed to marshal data", "error", err.Error())
		return &driverpb.FindDriverResponse{}, err
	}
	d.logger.Info("Sending Kafka message", slog.String("message", string(messageBytes)))

	err = d.kafkaProducer.SendMessage(ctx, kafka.Message{
		Value: messageBytes,
	})
	if err != nil {
		d.logger.Error("Failed to send message - Kafka", "error", err.Error())
		return &driverpb.FindDriverResponse{}, err
	}

	return &driverpb.FindDriverResponse{
		Driver:  selectedDriver,
		Success: true,
		Message: "Suitable driver found",
	}, nil

}

func (d *DriverGRPCService) GetAvailableDrivers(ctx context.Context, req *emptypb.Empty) (*driverpb.GetAvailableDriversResponse, error) {
	res, err := d.driverRepo.GetAvailableDrivers(ctx)
	if err != nil {
		d.logger.Error("failed to get available drivers", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	drivers := make([]*driverpb.Driver, 0, len(res))
	for _, driver := range res {
		drivers = append(drivers, &driverpb.Driver{
			DriverId: driver.ID,
			Name:     driver.Name,
			Phone:    driver.Phone,
			Status:   string(driver.Status),
			Vehicle: &driverpb.Vehicle{
				Model:        driver.Car,
				LicensePlate: driver.LicenseNumber,
			},
		})
	}
	return &driverpb.GetAvailableDriversResponse{
		Drivers: drivers,
	}, nil
}

func (d *DriverGRPCService) UpdateDriverStatus(ctx context.Context, req *driverpb.UpdateDriverStatusRequest) (*driverpb.UpdateDriverStatusResponse, error) {
	err := d.driverRepo.UpdateDriverStatus(ctx, int(req.DriverId), req.Status)
	if err != nil {
		d.logger.Error("failed to update driver status", slog.String("status", "error"), slogger.Err(err))
		return &driverpb.UpdateDriverStatusResponse{
			Success: false,
		}, nil
	}
	return &driverpb.UpdateDriverStatusResponse{
		Success: true,
	}, nil
}
