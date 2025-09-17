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
			ProductID:   stockItem.ProductId,
			ProductName: stockItem.ProductName,
			Quantity:    stockItem.Quantity,
			LastUpdated: stockItem.Time,
		}
	}

	return goodsItems
}

func ConvertOrderItemsToStock(goodsItems []*entity.GoodsItem) []*warehousepb.Stock {
	stockItems := make([]*warehousepb.Stock, len(goodsItems))

	for i, orderItem := range goodsItems {
		stockItems[i] = &warehousepb.Stock{
			ProductName: orderItem.ProductName,
			Quantity:    orderItem.Quantity,
			Price:       orderItem.Price,
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
			TotalPrice:  item.TotalPrice,
		}
	}
	return goodsItems
}

func ConvertOrderItemToWarehouseStockItem(orderItem []*orderpb.OrderItem, time int64) []*warehousepb.StockItem {
	stockItems := make([]*warehousepb.StockItem, len(orderItem))

	for i, item := range orderItem {
		stockItems[i] = &warehousepb.StockItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Time:        time,
		}
	}
	return stockItems
}

func ConvertGoodsItemSliceToOrderItemSlice(goodsItems []entity.GoodsItem) []*orderpb.OrderItem {
	if goodsItems == nil {
		return nil
	}

	orderItems := make([]*orderpb.OrderItem, len(goodsItems))
	for i, goodsItem := range goodsItems {
		orderItems[i] = &orderpb.OrderItem{
			ProductId:   goodsItem.ProductID,
			ProductName: goodsItem.ProductName,
			Price:       goodsItem.Price,
			Quantity:    goodsItem.Quantity,
			TotalPrice:  goodsItem.TotalPrice,
		}
	}
	return orderItems
}

// ConvertOrders converts a slice of orderpb.Order to a slice of entity.Order
func ConvertOrders(pbOrders []*orderpb.Order) []*entity.Order {
	var orders []*entity.Order
	for _, pbOrder := range pbOrders {
		orders = append(orders, ConvertOrder(pbOrder))
	}
	return orders
}

// ConvertOrder converts a single orderpb.Order to entity.Order
func ConvertOrder(pbOrder *orderpb.Order) *entity.Order {
	var createdAt int64
	if pbOrder.CreatedAt != nil {
		createdAt = pbOrder.CreatedAt.AsTime().Unix()
	}

	return &entity.Order{
		ID:              pbOrder.Id,
		UserID:          pbOrder.UserId,
		Status:          entity.OrderStatus(pbOrder.Status),
		DeliveryAddress: pbOrder.DeliveryAddress,
		Items:           ConvertOrderItems(pbOrder.Items),
		TotalAmount:     pbOrder.TotalAmount,
		DriverID:        &pbOrder.DriverId,
		CreatedAt:       createdAt,
	}
}

// ConvertOrderItems converts a slice of orderpb.OrderItem to a slice of entity.GoodsItem
func ConvertOrderItems(pbItems []*orderpb.OrderItem) []entity.GoodsItem {
	var items []entity.GoodsItem
	for _, pbItem := range pbItems {

		items = append(items, entity.GoodsItem{
			ProductID:   pbItem.ProductId,
			ProductName: pbItem.ProductName,
			Price:       pbItem.Price,
			Quantity:    pbItem.Quantity,
			TotalPrice:  pbItem.TotalPrice,
		})
	}
	return items
}

func ConvertStocksToGoodsItems(stocks []*warehousepb.Stock) []*entity.GoodsItem {
	goodsItems := make([]*entity.GoodsItem, 0, len(stocks))
	for _, stock := range stocks {
		goodsItem := &entity.GoodsItem{
			ProductID:   stock.ProductId,
			ProductName: stock.ProductName,
			Price:       stock.Price,
			Quantity:    stock.Quantity,
		}
		goodsItems = append(goodsItems, goodsItem)
	}
	return goodsItems
}
