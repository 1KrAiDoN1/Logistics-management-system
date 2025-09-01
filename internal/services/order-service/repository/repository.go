package repository

import (
	"context"
	"fmt"
	"logistics/internal/shared/entity"

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

}

func (o *OrderRepository) CompleteDelivery(ctx context.Context, userID, orderID int64) error {
	query := `UPDATE orders SET status = 'delivered' WHERE id = $1 AND user_id = $2`
	_, err := o.pool.Exec(ctx, query, orderID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (o *OrderRepository) GetDeliveriesByUser(ctx context.Context, userID int64) ([]*entity.Order, error) {
	query := `SELECT id, user_id, status, total_amount, delivery_address, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := o.pool.Query(ctx, query, userID)
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

func (o *OrderRepository) GetOrderDetails(ctx context.Context, userID, orderID int64) (*entity.Order, error) {
	query := `SELECT id, user_id, status, total_amount, delivery_address, created_at, driver_id FROM orders WHERE id = $1`
	row := o.pool.QueryRow(ctx, query, orderID)

	var order entity.Order
	err := row.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.DeliveryAddress, &order.CreatedAt, &order.DriverID)
	if err != nil {
		return nil, err
	}

	// Fetch order items
	itemsQuery := `SELECT product_id, product_name, price, quantity FROM order_items WHERE order_id = $1`
	itemsRows, err := o.pool.Query(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	var items []entity.GoodsItem
	for itemsRows.Next() {
		var item entity.GoodsItem
		err := itemsRows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.Quantity)
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

func (o *OrderRepository) UpdateOrderStatus(ctx context.Context, userID, orderID int64, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2 AND user_id = $3`
	_, err := o.pool.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}
	return nil
}
