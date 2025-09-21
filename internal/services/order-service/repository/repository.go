package repository

import (
	"context"
	"fmt"
	"logistics/internal/shared/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (o *OrderRepository) CreateOrder(ctx context.Context, order *entity.Order) (int64, error) {
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO orders (user_id, driver_id, status, delivery_address, total_amount, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var orderID int64
	err = tx.QueryRow(ctx, query,
		order.UserID,
		0,
		order.Status,
		order.DeliveryAddress,
		order.TotalAmount,
		order.CreatedAt,
	).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert order: %w", err)
	}

	if len(order.Items) > 0 {
		itemsQuery := `INSERT INTO order_items (order_id, product_id, product_name, price, quantity, total_price, last_updated) VALUES ($1, $2, $3, $4, $5, $6, $7)`

		batch := &pgx.Batch{}
		for _, item := range order.Items {
			batch.Queue(itemsQuery,
				orderID,
				item.ProductID,
				item.ProductName,
				item.Price,
				item.Quantity,
				item.TotalPrice,
				order.CreatedAt,
			)
		}

		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return 0, fmt.Errorf("failed to insert order items: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orderID, nil

}

func (o *OrderRepository) CompleteDelivery(ctx context.Context, userID, orderID int64) (int64, error) {
	query := `UPDATE orders SET status = 'delivered' WHERE id = $1 AND user_id = $2 RETURNING driver_id`
	var driverID int64
	err := o.pool.QueryRow(ctx, query, orderID, userID).Scan(&driverID)
	if err != nil {
		return 0, err
	}
	return driverID, nil
}

func (o *OrderRepository) CheckDeliveryStatus(ctx context.Context, userID, orderID int64) (string, error) {
	query := `SELECT status FROM orders WHERE id = $1 AND user_id = $2`
	var status string
	err := o.pool.QueryRow(ctx, query, orderID, userID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}
func (o *OrderRepository) GetDeliveriesByUser(ctx context.Context, userID int64) ([]*entity.Order, error) {
	query := `SELECT id, user_id, status, total_amount, delivery_address, created_at FROM orders WHERE user_id = $1 AND status = $2 ORDER BY created_at DESC`
	rows, err := o.pool.Query(ctx, query, userID, entity.StatusInProgress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.DeliveryAddress, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *OrderRepository) GetOrderItemInfo(ctx context.Context, productName string) (int32, float64, error) {
	query := `SELECT product_id, price FROM warehouse_stock WHERE product_name = $1`
	var product_id int32
	var price float64
	err := o.pool.QueryRow(ctx, query, productName).Scan(&product_id, &price)
	if err != nil {
		return 0, 0, err
	}
	return product_id, price, nil
}

func (o *OrderRepository) GetOrderDetails(ctx context.Context, userID, orderID int64) (*entity.Order, error) {
	query := `SELECT id, user_id, status, total_amount, delivery_address, created_at, driver_id FROM orders WHERE id = $1`
	row := o.pool.QueryRow(ctx, query, orderID)

	var order entity.Order
	err := row.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.DeliveryAddress, &order.CreatedAt, &order.DriverID)
	if err != nil {
		return nil, err
	}

	// Fetch order items
	itemsQuery := `SELECT product_id, product_name, price, quantity, total_price FROM order_items WHERE order_id = $1`
	itemsRows, err := o.pool.Query(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	var items []entity.GoodsItem
	for itemsRows.Next() {
		var item entity.GoodsItem
		err := itemsRows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.Quantity, &item.TotalPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := itemsRows.Err(); err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

func (o *OrderRepository) GetOrdersByUser(ctx context.Context, userID int64) ([]*entity.Order, error) {
	// Начинаем транзакцию
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Всегда откатываем, если не подтвердили

	// Получаем основные данные заказов
	query := `SELECT id, user_id, status, total_amount, delivery_address, created_at, driver_id FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order
	var orderIDs []int64 // Сохраняем ID заказов для batch-запроса товаров

	for rows.Next() {
		var order entity.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.DeliveryAddress, &order.CreatedAt, &order.DriverID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
		orderIDs = append(orderIDs, order.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	// Если заказов нет, сразу возвращаем результат
	if len(orders) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return orders, nil
	}

	// Получаем все товары для всех заказов одним запросом
	itemsQuery := `SELECT order_id, product_id, product_name, price, quantity FROM order_items WHERE order_id = ANY($1) ORDER BY order_id, product_id`
	itemsRows, err := tx.Query(ctx, itemsQuery, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer itemsRows.Close()

	// Создаем мапу для группировки товаров по ID заказа
	itemsByOrder := make(map[int64][]entity.GoodsItem)

	for itemsRows.Next() {
		var item entity.GoodsItem
		var orderID int64
		err := itemsRows.Scan(&orderID, &item.ProductID, &item.ProductName, &item.Price, &item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		itemsByOrder[orderID] = append(itemsByOrder[orderID], item)
	}

	if err := itemsRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order item rows: %w", err)
	}

	// Распределяем товары по заказам
	for _, order := range orders {
		if items, exists := itemsByOrder[order.ID]; exists {
			order.Items = items
		} else {
			order.Items = []entity.GoodsItem{} // Пустой слайс вместо nil
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orders, nil
}

func (o *OrderRepository) UpdateOrderStatus(ctx context.Context, userID, orderID int64, driverID int64, status string) error {
	query := `UPDATE orders SET status = $1, driver_id = $2 WHERE id = $3 AND user_id = $4`
	_, err := o.pool.Exec(ctx, query, status, driverID, orderID, userID)
	if err != nil {
		return err
	}
	return nil
}
