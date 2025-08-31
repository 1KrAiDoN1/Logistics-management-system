package utils

import (
	orderpb "logistics/api/protobuf/order_service"
	warehousepb "logistics/api/protobuf/warehouse_service"
	"logistics/internal/shared/entity"
)

func ConvertStockItemsToOrderItems(stockItems []*warehousepb.StockItem) []*entity.GoodsItem {
	goodsItems := make([]*entity.GoodsItem, len(stockItems))

	for i, stockItem := range stockItems {
		goodsItems[i] = &entity.GoodsItem{
			ProductName: stockItem.ProductName,
			Quantity:    stockItem.Quantity,
		}
	}

	return goodsItems
}

func ConvertOrderItemsToStock(goodsItems []*entity.GoodsItem) []*warehousepb.Stock {
	stockItems := make([]*warehousepb.Stock, len(goodsItems))

	for i, orderItem := range goodsItems {
		stockItems[i] = &warehousepb.Stock{
			ProductId:   orderItem.ProductID,
			ProductName: orderItem.ProductName,
			Quantity:    orderItem.Quantity,
		}
	}

	return stockItems
}

func ConvertOrderItemToGoodsItem(orderItem []*orderpb.OrderItem) []entity.GoodsItem {
	goodsItems := make([]entity.GoodsItem, len(orderItem))

	for i, item := range orderItem {
		goodsItems[i] = entity.GoodsItem{
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}
	return goodsItems
}
