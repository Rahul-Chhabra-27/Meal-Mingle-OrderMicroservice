package main

import (
	"context"
	"fmt"
	"order-microservice/config"
	"order-microservice/model"
	orderpb "order-microservice/proto/order"

	"google.golang.org/grpc/codes"
)

func (*OrderService) AddOrder(ctx context.Context, response *orderpb.AddOrderRequest) (*orderpb.AddOrderItemResponse, error) {
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		fmt.Println("Failed to get user email from context")
		return &orderpb.AddOrderItemResponse{Message: "", Error: "Internal Server Error", StatusCode: int64(codes.Internal)}, nil
	}
	if len(response.OrderItems) == 0 || !config.ValidateOrderFields(response.RestaurantName, response.OrderTotalPrice) {
		return &orderpb.AddOrderItemResponse{
			Message:    "",
			StatusCode: int64(codes.InvalidArgument),
			Error:      "Invalid order fields",
		}, nil
	}
	// fetching the user from user email.
	var user model.User
	user.Email = userEmail
	userDBConnector, err := config.GetUserConnector(config.GoDotEnvVariable("DB_CONFIG"))
	if err != nil {
		return &orderpb.AddOrderItemResponse{
			Message:    "",
			StatusCode: int64(codes.Internal),
			Error:      "Failed to connect to database",
		}, nil
	}
	userDBConnector.Where("email = ?", user.Email).First(&user)
	// fetching the restaurant from restaurant name.
	var restaurant model.Restaurant
	restaurant.Name = response.RestaurantName
	restaurantDBConnector, err := config.GetRestaurantConnector(config.GoDotEnvVariable("DB_CONFIG"))
	if err != nil {
		return &orderpb.AddOrderItemResponse{
			Message:    "",
			StatusCode: int64(codes.Internal),
			Error:      "Failed to connect to database",
		}, nil
	}
	restaurantDBConnector.Where("name = ?", restaurant.Name).First(&restaurant)

	// creating the order.
	var order model.Order
	order.RestaurantId = restaurant.ID
	order.TotalPrice = response.OrderTotalPrice
	order.UserId = user.ID
	order.Status = "Pending"
	order.Discount = 0;
	// insert the order into the database.
	primaryKey := orderDBConnector.Create(&order)
	if primaryKey.Error != nil {
		return &orderpb.AddOrderItemResponse{
			Message:    "Failed to add order",
			StatusCode: int64(codes.Internal),
			Error:      primaryKey.Error.Error(),
		}, nil
	}
	// insert all the orderItems and using ordrid as foreign key.
	for _, orderItem := range response.OrderItems {
		var orderItemModel model.OrderItem
		orderItemModel.OrderId = order.ID
		orderItemModel.Name = orderItem.OrderItemName
		orderItemModel.Price = orderItem.OrderItemPrice
		orderItemModel.Quantity = orderItem.OrderItemQuantity
		primaryKey := orderItemDBConnector.Create(&orderItemModel)
		if primaryKey.Error != nil {
			return &orderpb.AddOrderItemResponse{
				Message:    "Failed to add order",
				StatusCode: int64(codes.Internal),
				Error:      primaryKey.Error.Error(),
			}, nil
		}
	}
	Caller(int(order.TotalPrice))
	return &orderpb.AddOrderItemResponse{
		Message:    "Order added successfully",
		StatusCode: 200,
	}, nil
}
