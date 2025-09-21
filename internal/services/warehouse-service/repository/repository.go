package repository

import (
	"context"
	"fmt"
	"logistics/internal/shared/entity"

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

	productNames := make([]string, 0, len(orders))
	for _, item := range orders {
		productNames = append(productNames, item.ProductName)
	}

	// Получаем текущие остатки по именам продуктов
	query := `SELECT product_name, quantity FROM warehouse_stock WHERE product_name = ANY($1)`

	rows, err := w.pool.Query(ctx, query, productNames)
	if err != nil {
		return false, fmt.Errorf("failed to get stock: %w", err)
	}
	defer rows.Close()

	// Создаем мапу остатков по именам продуктов
	stock := make(map[string]int32)
	for rows.Next() {
		var productName string
		var quantity int32
		if err := rows.Scan(&productName, &quantity); err != nil {
			return false, err
		}
		stock[productName] = quantity
	}

	// Проверяем достаточность остатков
	for _, item := range orders {
		available, exists := stock[item.ProductName]
		if !exists {
			return false, nil // Товар не найден на складе
		}
		if available < item.Quantity {
			return false, nil // Недостаточно товара
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

	// Получаем текущие количества товаров для проверки
	checkQuery := `SELECT quantity FROM warehouse_stock WHERE product_id = $1`

	updateQuery := `UPDATE warehouse_stock SET quantity = quantity - $1, last_updated = $2 WHERE product_id = $3`

	// Выполняем обновление для каждого товара в транзакции
	for _, item := range items {
		// Проверяем, достаточно ли товара на складе
		var currentQuantity int
		err := tx.QueryRow(ctx, checkQuery, item.ProductID).Scan(&currentQuantity)
		if err != nil {
			return fmt.Errorf("failed to get current quantity for product %d: %w", item.ProductID, err)
		}

		// Вычитаем количество заказанного товара
		_, err = tx.Exec(ctx, updateQuery,
			item.Quantity,    // количество для вычитания
			item.LastUpdated, // время обновления
			item.ProductID,   // ID товара
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
