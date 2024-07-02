package main

import (
	"context"
	"order-microservice/config"
	"order-microservice/model"
	orderpb "order-microservice/proto/order"

	"go.uber.org/zap"
)

func (*OrderService) ChangeOrderStatus(ctx context.Context, request *orderpb.ChangeOrderStatusRequest) (*orderpb.ChangeOrderStatusResponse, error) {
	// get the user mail from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		logger.Error("Failed to get user email from context")
		return &orderpb.ChangeOrderStatusResponse{
			Data:       nil,
			Message:    "Failed to get user email from context",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}

	// Validate the request
	if request.OrderId == "" || request.OrderStatus == "" {
		logger.Warn("Invalid request fields", zap.String("userEmail", userEmail))
		return &orderpb.ChangeOrderStatusResponse{
			Data:       nil,
			Message:    "Invalid request fields",
			StatusCode: StatusBadRequest,
			Error:      "Bad Request",
		}, nil
	}
	// get the user from the database
	var user model.User
	userDBConnector, userDbErr := config.GetUserConnector()
	if userDbErr != nil {
		logger.Error("Failed to connect to user database", zap.Error(userDbErr))
		return &orderpb.ChangeOrderStatusResponse{
			Data:       nil,
			Message:    "Server facing issues",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}
	if err := userDBConnector.Where("email = ?", userEmail).First(&user).Error; err != nil {
		logger.Warn("User not found", zap.String("userEmail", userEmail))
		return &orderpb.ChangeOrderStatusResponse{
			Data:       nil,
			Message:    "User not found",
			StatusCode: StatusNotFound,
			Error:      "Not Found",
		}, nil
	}

	// check if the order exists
	var order model.Order
	orderErr := orderDBConnector.Where("id = ?", request.OrderId).First(&order).Error

	if orderErr != nil || order.UserId != user.ID {
		logger.Warn("Unauthorized attempt to change order status",
			zap.String("userEmail", userEmail),
			zap.String("orderID", request.OrderId))
		return &orderpb.ChangeOrderStatusResponse{
			Data:       nil,
			Message:    "Not Authorized to change the order status",
			StatusCode: StatusUnauthorized,
			Error:      "Unauthorized",
		}, nil
	}
	// change the status of the order
	order.Status = request.OrderStatus
	saveErr := orderDBConnector.Save(&order).Error

	if saveErr != nil {
		logger.Error("Failed to save order status", zap.String("orderID", request.OrderId), zap.Error(saveErr))
		return &orderpb.ChangeOrderStatusResponse{
			Data:       nil,
			Message:    "Failed to change order status",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}
	logger.Info("Order status changed successfully",
		zap.String("orderID", request.OrderId), zap.String("newStatus", request.OrderStatus))
	// return the response
	return &orderpb.ChangeOrderStatusResponse{
		Data: &orderpb.ChangeOrderStatusResponseData{
			OrderId:     request.OrderId,
			OrderStatus: request.OrderStatus,
		},
		Message:    "Order status changed successfully",
		StatusCode: StatusOK,
		Error:      "",
	}, nil
}
