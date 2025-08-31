package repository

import (
	"context"
	"fmt"
	"logistics/internal/shared/entity"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WarehouseRepository struct {
	pool *pgxpool.Pool
}

func NewWarehouseRepository(pool *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{
		pool: pool,
	}
}

func (w *WarehouseRepository) CheckStockAvailability(ctx context.Context, orders []*entity.GoodsItem) (bool, error) {
	if len(orders) == 0 {
		return true, nil
	}

	// Собираем ID товаров для проверки
	productIDs := make([]int64, 0, len(orders))
	for _, item := range orders {
		productIDs = append(productIDs, item.ProductID)
	}

	// Получаем текущие остатки
	query := `
        SELECT product_id, quantity 
        FROM warehouse_stock 
        WHERE product_id = ANY($1)`

	rows, err := w.pool.Query(ctx, query, productIDs)
	if err != nil {
		return false, fmt.Errorf("failed to get stock: %w", err)
	}
	defer rows.Close()

	// Создаем мапу остатков
	stock := make(map[int64]int32)
	for rows.Next() {
		var productID int64
		var quantity int32
		if err := rows.Scan(&productID, &quantity); err != nil {
			return false, err
		}
		stock[productID] = quantity
	}

	// Проверяем достаточность остатков
	for _, item := range orders {
		available, exists := stock[item.ProductID]
		if !exists || available < item.Quantity {
			return false, nil
		}
	}

	return true, nil
}

func (w *WarehouseRepository) GetWarehouseStock(ctx context.Context) ([]*entity.GoodsItem, error) {
	query := `SELECT product_id, product_name, quantity, price, last_updated FROM warehouse_stock WHERE quantity > 0`

	rows, err := w.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query warehouse stock: %w", err)
	}
	defer rows.Close()

	var items []*entity.GoodsItem

	for rows.Next() {
		item := &entity.GoodsItem{}

		err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.Price,
			&item.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan warehouse item: %w", err)
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warehouse stock rows: %w", err)
	}

	return items, nil
}

func (w *WarehouseRepository) UpdateStock(ctx context.Context, items []*entity.GoodsItem) error {
	if len(items) == 0 {
		return nil
	}

	// Начинаем транзакцию
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	currentTime := time.Now().Unix()
	query := `
        UPDATE warehouse_stock 
        SET price = $1, quantity = $2, last_updated = $3
        WHERE product_id = $4
    `

	// Выполняем обновление для каждого товара в транзакции
	for _, item := range items {
		item.LastUpdated = currentTime

		_, err := tx.Exec(ctx, query,
			item.Price,
			item.Quantity,
			item.LastUpdated,
			item.ProductID,
		)
		if err != nil {
			return fmt.Errorf("failed to update stock for product %d: %w", item.ProductID, err)
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
