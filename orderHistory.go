package main

import (
	"context"
	orderpb "order-microservice/proto/order"
	"order-microservice/model"
	"order-microservice/config"
)

func (*OrderService) OrderHistory(ctx context.Context, resquest *orderpb.OrderHistoryRequest) (*orderpb.OrderHistoryResponse, error) {
	// fetch the user email from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		return &orderpb.OrderHistoryResponse{
			Message:    "",
			StatusCode: 500,
			Error:      "Internal Server Error",
		}, nil
	}
	// fetch the user from the user email
	var user model.User
	user.Email = userEmail
	userDBConnector, err := config.GetUserConnector(config.GoDotEnvVariable("DB_CONFIG"))
	if err != nil {
		return &orderpb.OrderHistoryResponse{
			Message:    "",
			StatusCode: 500,
			Error:      "Failed to connect to database",
		}, nil
	}
	userDBConnector.Where("email = ?", user.Email).First(&user)
	restaurantDBConnector, err := config.GetRestaurantConnector(config.GoDotEnvVariable("DB_CONFIG"))
	if err != nil {
		return &orderpb.OrderHistoryResponse{
			Message:    "",
			StatusCode: 500,
			Error:      "Failed to connect to database",
		}, nil
	}

	// fetch the orders from the user id
	var orders []model.Order
	orderDBConnector.Where("user_id = ?", user.ID).Find(&orders)
	// create the response
	var orderHistoryResponse orderpb.OrderHistoryResponse
	for _, order := range orders {
		// for every order we are fetching the order items
		var orderItems []model.OrderItem
		var restaurant model.Restaurant
		orderItemDBConnector.Where("order_id = ?", order.ID).Find(&orderItems)
		restaurantDBConnector.Where("id = ?", order.RestaurantId).First(&restaurant)
		var orderItemsResponse []*orderpb.OrderItem
		for _, orderItem := range orderItems {
			orderItemsResponse = append(orderItemsResponse, &orderpb.OrderItem{
				OrderItemName: orderItem.Name,
				OrderItemPrice: orderItem.Price,
				OrderItemQuantity: orderItem.Quantity,
			})
		}
		orderHistoryResponse.Orders = append(orderHistoryResponse.Orders, &orderpb.Order{
			OrderItems: orderItemsResponse,
			OrderTotalPrice: order.TotalPrice,
			RestaurantName: restaurant.Name,
		})
		orderHistoryResponse.Message = "Successfully fetched order history"
		orderHistoryResponse.StatusCode = 200
	}
	return &orderHistoryResponse, nil
}
