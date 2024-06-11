package main

import (
	"context"
	"fmt"
	"order-microservice/config"
	"order-microservice/model"
	orderpb "order-microservice/proto/order"
	"strconv"

	"google.golang.org/grpc/codes"
)

func (*OrderService) AddOrder(ctx context.Context, request *orderpb.AddOrderRequest) (*orderpb.AddOrderResponse, error) {
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		fmt.Println("Failed to get user email from context")
		return &orderpb.AddOrderResponse{Message: "", Error: "Internal Server Error", StatusCode: int64(500)}, nil
	}
	if len(request.OrderItems) == 0 || !config.ValidateOrderFields(request.ShippingAddress, request.RestaurantName, request.OrderTotalPrice) {
		return &orderpb.AddOrderResponse{
			Message:    "",
			StatusCode: int64(400),
			Error:      "Invalid order fields",
		}, nil
	}
	// fetching the user from user email.
	var user model.User
	user.Email = userEmail
	userDBConnector, err := config.GetUserConnector(config.GoDotEnvVariable("DB_CONFIG"))
	if err != nil {
		fmt.Println("Failed to connect to database")
		return &orderpb.AddOrderResponse{
			Message:    "",
			StatusCode: int64(codes.Internal),
			Error:      "Internal Server Error",
		}, nil
	}
	userDBConnector.Where("email = ?", user.Email).First(&user)
	// fetching the restaurant from restaurant name.
	var restaurant model.Restaurant
	restaurant.Name = request.RestaurantName
	restaurantDBConnector, err := config.GetRestaurantConnector(config.GoDotEnvVariable("DB_CONFIG"))
	if err != nil {
		fmt.Println("Failed to connect to database")
		return &orderpb.AddOrderResponse{
			Message:    "",
			StatusCode: int64(500),
			Error:      "Internal Server Error",
		}, nil
	}
	restaurantDBConnector.Where("name = ?", restaurant.Name).First(&restaurant)
	// payment gateway logic goes here.
	result := Caller(int(request.OrderTotalPrice))
	if result != "Payment Successfull" {
		return &orderpb.AddOrderResponse{
			Message:    "",
			StatusCode: int64(401),
			Error:      "Payment Required, Failed to add order",
		}, nil
	}
	// creating the order.
	var order model.Order
	order.RestaurantId = restaurant.ID
	order.TotalPrice = request.OrderTotalPrice
	order.UserId = user.ID
	order.Status = "Delivered"
	order.Discount = 0
	order.ShippingAddress = request.ShippingAddress
	// insert the order into the database.
	primaryKey := orderDBConnector.Create(&order)
	if primaryKey.Error != nil {
		fmt.Println("Failed to add order")
		return &orderpb.AddOrderResponse{
			Message:    "",
			StatusCode: int64(codes.Internal),
			Error:      "Internal Server Error",
		}, nil
	}
	// insert all the orderItems and using ordrid as foreign key.
	for _, orderItem := range request.OrderItems {
		var orderItemModel model.OrderItem
		orderItemModel.OrderId = order.ID
		orderItemModel.Name = orderItem.OrderItemName
		orderItemModel.Price = orderItem.OrderItemPrice
		orderItemModel.Quantity = orderItem.OrderItemQuantity
		primaryKey := orderItemDBConnector.Create(&orderItemModel)

		if primaryKey.Error != nil {
			fmt.Println("Failed to add order item")
			return &orderpb.AddOrderResponse{
				Error:      "Failed to add order",
				StatusCode: int64(codes.Internal),
				Message:    "",
			}, nil
		}
	}
	return &orderpb.AddOrderResponse{
		Message:    "Order added successfully",
		StatusCode: 200,
		Error:      "",
		Data: &orderpb.Data{
			Order: []*orderpb.Order{
				{
					OrderId:         strconv.FormatUint(uint64(order.ID), 10),
					OrderItems:      request.OrderItems,
					OrderTotalPrice: order.TotalPrice,
					RestaurantName:  restaurant.Name,
					OrderStatus:     "Processing🍔",
					ShippingAddress: request.ShippingAddress,
				},
			},
		},
	}, nil
}
