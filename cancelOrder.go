package main

import (
	"context"
	"order-microservice/config"
	"order-microservice/model"
	orderpb "order-microservice/proto/order"

	"go.uber.org/zap"
)

func (*OrderService) CancelOrder(ctx context.Context, request *orderpb.CancelOrderRequest) (response *orderpb.CancelOrderResponse, err error) {
	// userEmail from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		logger.Error("Failed to get user email from context")
		return &orderpb.CancelOrderResponse{
			Message:    "Failed to get user email from context",
			Error:      "Internal Server Error",
			StatusCode: StatusInternalServerError,
		}, nil
	}
	// Validate the fields
	if request.OrderId == "" {
		logger.Warn("Invalid Order ID", zap.String("userEmail", userEmail))
		return &orderpb.CancelOrderResponse{
			Message:    "Invalid Order ID",
			Error:      "Invalid fields, make sure to use mentioned format.",
			StatusCode: StatusBadRequest,
		}, nil
	}

	//check if the order exists
	var order model.Order
	orderID := request.OrderId

	res := orderDBConnector.Where("id = ?", orderID).First(&order)
	if res.Error != nil {
		logger.Warn("Order not found", zap.String("orderID", orderID))
		return &orderpb.CancelOrderResponse{
			Message:    "Order not found with the given ID",
			Error:      "Not Found",
			StatusCode: StatusNotFound,
		}, nil
	}

	// check if the order exists or order belongs to the user
	var user model.User
	user.Email = userEmail
	userDBConnector, userDbErr := config.GetUserConnector()
	if userDbErr != nil {
		logger.Error("Failed to connect to user database", zap.Error(userDbErr))
		return &orderpb.CancelOrderResponse{
			Message:    "Server facing issues",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}

	userDBConnector.Where("email = ?", user.Email).First(&user)
	if order.UserId != user.ID {
		logger.Warn("Unauthorized order cancellation attempt", zap.String("userEmail", userEmail), zap.String("orderID", orderID))
		return &orderpb.CancelOrderResponse{
			Message:    "Order does not belong to the user, you are not authorized to cancel this order.",
			Error:      "Unauthorized",
			StatusCode: StatusUnauthorized,
		}, nil
	}

	// delete the user..
	res = orderDBConnector.Delete(&order)
	if res.Error != nil {
		logger.Error("Failed to delete the order", zap.String("orderID", orderID))
		return &orderpb.CancelOrderResponse{
			Message:    "Failed to delete the order",
			Error:      "Internal server error",
			StatusCode: StatusInternalServerError,
		}, nil
	}
	logger.Info("Order cancelled successfully", zap.String("orderID", orderID))
	return &orderpb.CancelOrderResponse{
		Data:       &orderpb.CancelOrderData{OrderId: request.OrderId},
		Message:    "Order Cancelled Successfully",
		Error:      "",
		StatusCode: StatusOK,
	}, nil
}
