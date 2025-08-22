// cmd/demo/main.go - Демонстрация полного флоу создания заказа
package main

import (
	"context"
	"fmt"
	"time"
)

// Демонстрация полного процесса создания заказа
func main() {
	fmt.Println("=== LOGISTICS SYSTEM DEMO ===")
	fmt.Println("Демонстрация процесса создания заказа")

	// 1. Пользователь делает запрос на создание заказа
	orderRequest := `{
		"user_id": 12345,
		"delivery_address": "Москва, ул. Тверская, д. 1",
		"pickup_address": "Склад №1, Московская область, Химки",
		"items": [
			{
				"product_id": 1001,
				"quantity": 2,
				"price": 1500.00
			},
			{
				"product_id": 1002,
				"quantity": 1,
				"price": 3000.00
			}
		],
		"priority": "normal",
		"scheduled_at": null,
		"notes": "Доставить до 18:00"
	}`

	fmt.Printf("\n1. Пользователь создает заказ:\n%s\n", orderRequest)

	// 2. API Gateway обрабатывает запрос
	fmt.Println("\n2. API Gateway получает запрос и начинает обработку:")

	ctx := context.Background()

	// Шаг 2.1: Проверка наличия на складе
	fmt.Println("\n   2.1. Вызов warehouse-service: Проверка наличия товаров")
	stockCheckResult := simulateStockCheck(ctx)
	fmt.Printf("   Результат: %+v\n", stockCheckResult)

	if !stockCheckResult.Available {
		fmt.Println("   ❌ Недостаточно товара на складе. Заказ отклонен.")
		return
	}

	// Шаг 2.2: Резервирование товаров
	fmt.Println("\n   2.2. Вызов warehouse-service: Резервирование товаров")
	reservationResult := simulateStockReservation(ctx)
	fmt.Printf("   Резервирование: %s до %s\n",
		reservationResult.ReservationID,
		reservationResult.ExpiresAt.Format("15:04:05"))

	// Шаг 2.3: Создание заказа
	fmt.Println("\n   2.3. Вызов order-service: Создание заказа")
	orderResult := simulateOrderCreation(ctx, reservationResult.ReservationID)
	fmt.Printf("   Заказ создан: ID=%d, Статус=%s\n",
		orderResult.OrderID, orderResult.Status)

	// Шаг 2.4: Асинхронная генерация маршрута
	fmt.Println("\n3. Асинхронная обработка (в фоне):")
	fmt.Println("   3.1. Вызов route-service: Генерация маршрута")

	time.Sleep(1 * time.Second) // Имитируем время обработки
	routeResult := simulateRouteGeneration(ctx, orderResult.OrderID)
	fmt.Printf("   Маршрут создан: ID=%d, Расстояние=%.1f км, Время=%.0f мин\n",
		routeResult.RouteID, routeResult.Distance, routeResult.EstimatedTime.Minutes())

	// Обновляем статус заказа
	simulateOrderStatusUpdate(ctx, orderResult.OrderID, "route_ready", "Маршрут построен")

	// Шаг 2.5: Поиск и назначение водителя
	fmt.Println("\n   3.2. Вызов driver-service: Поиск подходящего водителя")
	time.Sleep(2 * time.Second) // Имитируем время поиска

	driverResult := simulateFindDriver(ctx, orderResult.OrderID, routeResult)
	if driverResult == nil {
		fmt.Println("   ⚠️ Подходящий водитель не найден. Заказ в очереди.")
		simulateOrderStatusUpdate(ctx, orderResult.OrderID, "pending_driver", "Ожидание водителя")
		return
	}

	fmt.Printf("   Водитель найден: %s (ID=%d), Рейтинг=%.1f, Расстояние=%.1f км\n",
		driverResult.Name, driverResult.ID, driverResult.Rating, driverResult.DistanceToPickup)

	// Назначаем водителя
	simulateDriverAssignment(ctx, orderResult.OrderID, driverResult.ID)
	simulateOrderStatusUpdate(ctx, orderResult.OrderID, "assigned",
		fmt.Sprintf("Назначен водитель: %s", driverResult.Name))

	fmt.Println("\n4. Заказ успешно создан и назначен водителю!")

	// Показываем итоговое состояние заказа
	finalOrder := simulateGetOrderDetails(ctx, orderResult.OrderID)
	fmt.Printf("\n=== ИТОГОВОЕ СОСТОЯНИЕ ЗАКАЗА ===\n")
	fmt.Printf("ID заказа: %d\n", finalOrder.OrderID)
	fmt.Printf("Статус: %s\n", finalOrder.Status)
	fmt.Printf("Водитель: %s (%s)\n", finalOrder.DriverName, finalOrder.DriverPhone)
	fmt.Printf("Транспорт: %s %s\n", finalOrder.VehicleModel, finalOrder.VehiclePlate)
	fmt.Printf("Маршрут: %.1f км, ~%.0f мин\n", finalOrder.Distance, finalOrder.EstimatedTime.Minutes())
	fmt.Printf("Сумма заказа: %.2f руб\n", finalOrder.TotalAmount)

	fmt.Println("\n5. Симуляция процесса доставки...")
	simulateDeliveryProcess(ctx, orderResult.OrderID, driverResult.ID)

	fmt.Println("\n✅ Демонстрация завершена!")
}

// Симуляция проверки склада
func simulateStockCheck(ctx context.Context) *StockCheckResult {
	fmt.Println("      - Проверка товара 1001: количество 2")
	fmt.Println("      - Склад №1: доступно 15 шт")
	fmt.Println("      - Проверка товара 1002: количество 1")
	fmt.Println("      - Склад №1: доступно 8 шт")

	return &StockCheckResult{
		Available: true,
		Items: []StockItem{
			{ProductID: 1001, Requested: 2, Available: 15, WarehouseName: "Склад №1"},
			{ProductID: 1002, Requested: 1, Available: 8, WarehouseName: "Склад №1"},
		},
		Message: "Все товары в наличии",
	}
}

// Симуляция резервирования товаров
func simulateStockReservation(ctx context.Context) *ReservationResult {
	fmt.Println("      - Резервирование товара 1001: 2 шт на складе №1")
	fmt.Println("      - Резервирование товара 1002: 1 шт на складе №1")

	return &ReservationResult{
		ReservationID: "res_1703123456_789123",
		ExpiresAt:     time.Now().Add(5 * time.Minute),
		Status:        "active",
	}
}

// Симуляция создания заказа
func simulateOrderCreation(ctx context.Context, reservationID string) *OrderCreationResult {
	fmt.Println("      - Создание записи в БД заказов")
	fmt.Println("      - Сохранение товаров заказа")
	fmt.Println("      - Подтверждение резервирования:", reservationID)
	fmt.Println("      - Статус заказа: confirmed")

	return &OrderCreationResult{
		OrderID: 88001,
		Status:  "confirmed",
	}
}

// Симуляция генерации маршрута
func simulateRouteGeneration(ctx context.Context, orderID int64) *RouteResult {
	fmt.Println("      - Расчет оптимального маршрута от склада до адреса доставки")
	fmt.Println("      - Учет текущей дорожной обстановки")
	fmt.Println("      - Сохранение маршрута в БД")

	return &RouteResult{
		RouteID:       301,
		Distance:      27.5,
		EstimatedTime: 45 * time.Minute,
	}
}

// Симуляция поиска водителя
func simulateFindDriver(ctx context.Context, orderID int64, route *RouteResult) *DriverResult {
	fmt.Println("      - Поиск доступных водителей в радиусе 30 км")
	fmt.Println("      - Найдено 5 потенциальных водителей")
	fmt.Println("      - Анализ пригодности:")
	fmt.Println("        * Водитель Петров И. - рейтинг 4.8, расстояние 8 км, опыт 1200+ заказов")
	fmt.Println("        * Водитель Сидоров П. - рейтинг 4.6, расстояние 15 км, опыт 800+ заказов")
	fmt.Println("        * Водитель Иванов С. - рейтинг 4.3, расстояние 12 км, опыт 300+ заказов")
	fmt.Println("      - Выбран лучший водитель: Петров И.")

	return &DriverResult{
		ID:               2001,
		Name:             "Петров Иван",
		Phone:            "+7-999-123-45-67",
		Rating:           4.8,
		DistanceToPickup: 8.2,
		VehicleModel:     "ГАЗель NEXT",
		VehiclePlate:     "М123АВ77",
	}
}

// Симуляция назначения водителя
func simulateDriverAssignment(ctx context.Context, orderID, driverID int64) {
	fmt.Printf("      - Обновление статуса водителя %d: available -> busy\n", driverID)
	fmt.Printf("      - Назначение заказа %d водителю %d\n", orderID, driverID)
	fmt.Println("      - Отправка уведомления водителю")
	fmt.Println("      - Отправка уведомления клиенту")
}

// Симуляция обновления статуса заказа
func simulateOrderStatusUpdate(ctx context.Context, orderID int64, status, message string) {
	fmt.Printf("      - Обновление статуса заказа %d: %s (%s)\n", orderID, status, message)
}

// Симуляция получения деталей заказа
func simulateGetOrderDetails(ctx context.Context, orderID int64) *OrderDetails {
	return &OrderDetails{
		OrderID:       orderID,
		Status:        "assigned",
		DriverName:    "Петров Иван",
		DriverPhone:   "+7-999-123-45-67",
		VehicleModel:  "ГАЗель NEXT",
		VehiclePlate:  "М123АВ77",
		Distance:      27.5,
		EstimatedTime: 45 * time.Minute,
		TotalAmount:   6000.00,
	}
}

// Симуляция процесса доставки
func simulateDeliveryProcess(ctx context.Context, orderID, driverID int64) {
	fmt.Printf("   5.1. Водитель начинает движение к складу (статус: in_progress)\n")
	simulateOrderStatusUpdate(ctx, orderID, "in_progress", "Водитель направляется к складу")
	time.Sleep(1 * time.Second)

	fmt.Printf("   5.2. Водитель прибыл на склад, загружает товар\n")
	time.Sleep(1 * time.Second)

	fmt.Printf("   5.3. Водитель направляется к клиенту\n")
	time.Sleep(1 * time.Second)

	fmt.Printf("   5.4. Доставка выполнена (статус: delivered)\n")
	simulateOrderStatusUpdate(ctx, orderID, "delivered", "Заказ успешно доставлен")

	fmt.Printf("   5.5. Водитель освобожден (статус: available)\n")
	fmt.Printf("      - Обновление статуса водителя %d: busy -> available\n", driverID)
	fmt.Printf("      - Увеличение счетчика выполненных заказов\n")
	fmt.Printf("      - Обновление рейтинга водителя\n")
}

// Структуры для демонстрации
type StockCheckResult struct {
	Available bool        `json:"available"`
	Items     []StockItem `json:"items"`
	Message   string      `json:"message"`
}

type StockItem struct {
	ProductID     int64  `json:"product_id"`
	Requested     int32  `json:"requested"`
	Available     int32  `json:"available"`
	WarehouseName string `json:"warehouse_name"`
}

type ReservationResult struct {
	ReservationID string    `json:"reservation_id"`
	ExpiresAt     time.Time `json:"expires_at"`
	Status        string    `json:"status"`
}

type OrderCreationResult struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

type RouteResult struct {
	RouteID       int64         `json:"route_id"`
	Distance      float64       `json:"distance"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

type DriverResult struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Phone            string  `json:"phone"`
	Rating           float64 `json:"rating"`
	DistanceToPickup float64 `json:"distance_to_pickup"`
	VehicleModel     string  `json:"vehicle_model"`
	VehiclePlate     string  `json:"vehicle_plate"`
}

type OrderDetails struct {
	OrderID       int64         `json:"order_id"`
	Status        string        `json:"status"`
	DriverName    string        `json:"driver_name"`
	DriverPhone   string        `json:"driver_phone"`
	VehicleModel  string        `json:"vehicle_model"`
	VehiclePlate  string        `json:"vehicle_plate"`
	Distance      float64       `json:"distance"`
	EstimatedTime time.Duration `json:"estimated_time"`
	TotalAmount   float64       `json:"total_amount"`
}
