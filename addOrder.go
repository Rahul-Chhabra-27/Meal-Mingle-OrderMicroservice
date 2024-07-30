package main

import (
	"context"
	"order-microservice/config"
	"order-microservice/model"
	orderpb "order-microservice/proto/order"
	"strconv"

	"go.uber.org/zap"
)

func (*OrderService) AddOrder(ctx context.Context, request *orderpb.AddOrderRequest) (*orderpb.AddOrderResponse, error) {
	userEmail, emailCtxError := ctx.Value("userEmail").(string)
	userRole, roleCtxError := ctx.Value("userRole").(string)

	if !emailCtxError || !roleCtxError {
		logger.Error("Failed to get user email and role from context",
			zap.Bool("emailCtxError", emailCtxError), zap.Bool("roleCtxError", roleCtxError))
		return &orderpb.AddOrderResponse{
			Data:       nil,
			Message:    "Failed to get user mail and role from context",
			Error:      "Internal Server Error",
			StatusCode: StatusInternalServerError,
		}, nil
	}
	logger.Info("Received AddOrder request", 
	zap.String("userEmail", userEmail), zap.String("userRole", userRole))
	if userRole != model.UserRole {
		logger.Warn("Permission denied for non-user role", 
		zap.String("userRole", userRole))

		return &orderpb.AddOrderResponse{
			Data:       nil,
			Message:    "You do not have permission to perform this action. Only user can add an order",
			StatusCode: StatusForbidden,
			Error:      "Forbidden",
		}, nil
	}

	if len(request.OrderItems) == 0 || !config.ValidateOrderFields(request.ShippingAddress, request.RestaurantName, request.OrderTotalPrice) {
		logger.Warn("Invalid request fields", zap.String("userEmail", userEmail))
		return &orderpb.AddOrderResponse{
			Data:       nil,
			Message:    "Invalid request. Please check the fields and try again.",
			StatusCode: StatusBadRequest,
			Error:      "The request contains invalid or missing fields.",
		}, nil
	}
	// fetching the user from user email.
	var user model.User
	var restaurant model.Restaurant
	userDBConnector, userDbErr := config.GetUserConnector()
	restaurantDBConnector, restaurantDbErr := config.GetRestaurantConnector()
	if userDbErr != nil || restaurantDbErr != nil {
		logger.Error("Failed to connect to database",
			zap.Error(userDbErr), zap.Error(restaurantDbErr))
		return &orderpb.AddOrderResponse{
			Message:    "Server Facing Issues",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}
	logger.Info("Connected to databases")
	if err := userDBConnector.Where("email = ?", userEmail).First(&user).Error; err != nil {
		logger.Warn("User not found", zap.String("userEmail", userEmail))
		return &orderpb.AddOrderResponse{
			Message:    "User not found",
			StatusCode: StatusNotFound,
			Error:      "Resource not found",
		}, nil
	}

	restaurantError := restaurantDBConnector.Where("name = ?", request.RestaurantName).First(&restaurant).Error

	if restaurantError != nil {
		logger.Warn("Restaurant not found", zap.String("restaurantName", request.RestaurantName))
		return &orderpb.AddOrderResponse{
			Message:    "Restaurant not found",
			StatusCode: StatusNotFound,
			Error:      "Resource not found",
		}, nil
	}
	// payment gateway logic goes here.
	result := Caller(int(request.OrderTotalPrice))
	if result != "Payment Successfull" {
		logger.Warn("Payment failed", zap.String("userEmail", userEmail))
		return &orderpb.AddOrderResponse{
			Message:    "Payment Required. Failed to add order.",
			StatusCode: StatusBadRequest,
			Error:      "Payment Required, Failed to add order",
		}, nil
	}
	// creating the order.
	var order model.Order
	order.RestaurantId = restaurant.ID
	order.TotalPrice = request.OrderTotalPrice
	order.UserId = user.ID
	order.Status = "Pending"
	order.Discount = request.OrderDiscount
	order.ShippingAddress = request.ShippingAddress
	// insert the order into the database.
	orderDBConnector.Create(&order)
	logger.Info("Order created", zap.Uint("orderId", order.ID))

	// insert all the orderItems and using ordrid as foreign key.
	for _, orderItem := range request.OrderItems {
		var orderItemModel model.OrderItem
		orderItemModel.OrderId = order.ID
		orderItemModel.Name = orderItem.OrderItemName
		orderItemModel.Price = orderItem.OrderItemPrice
		orderItemModel.Quantity = orderItem.OrderItemQuantity

		orderItemDBConnector.Create(&orderItemModel)
		logger.Info("Order item created", zap.Uint("orderId", order.ID), zap.String("itemName", orderItem.OrderItemName))
	}
	logger.Info("Order added successfully", zap.Uint("orderId", order.ID))
	return &orderpb.AddOrderResponse{
		Message:    "Order added successfully",
		StatusCode: StatusOK,
		Error:      "",
		Data: &orderpb.Data{
			Order: []*orderpb.Order{
				{
					OrderId:         strconv.FormatUint(uint64(order.ID), 10),
					OrderItems:      request.OrderItems,
					OrderTotalPrice: order.TotalPrice,
					RestaurantName:  restaurant.Name,
					OrderStatus:     "Pending🍔",
					ShippingAddress: request.ShippingAddress,
				},
			},
		},
	}, nil
}
