package main

import (
	"context"
	"fmt"
	"order-microservice/config"
	"order-microservice/model"
	orderpb "order-microservice/proto/order"
)

func (*OrderService) CancelOrder(ctx context.Context, request *orderpb.CancelOrderRequest) (response *orderpb.CancelOrderResponse, err error) {
	// userEmail from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		fmt.Println("Failed to get user email from context")
		return &orderpb.CancelOrderResponse{Message: "", Error: "Internal Server Error", StatusCode: int64(500)}, nil
	}
	// Validate the fields
	if request.OrderId == "" {
		return &orderpb.CancelOrderResponse{Message: "", Error: "Invalid Order ID", StatusCode: int64(400)}, nil
	}

	//check if the order exists
	var order model.Order
	orderID := request.OrderId
	if err != nil {
		fmt.Println("Failed to connect to database")
		return &orderpb.CancelOrderResponse{Message: "", Error: "Internal Server Error", StatusCode: int64(500)}, nil
	}
	res := orderDBConnector.Where("id = ?", orderID).First(&order)
	if res.Error != nil {
		return &orderpb.CancelOrderResponse{Message: "", Error: "Order not found", StatusCode: int64(404)}, nil
	}

	// check if the order exists or order belongs to the user
	var user model.User
	user.Email = userEmail
	userDBConnector, err := config.GetUserConnector()
	userDBConnector.Where("email = ?", user.Email).First(&user)
	if order.UserId != user.ID {
		return &orderpb.CancelOrderResponse{Message: "", Error: "Order does not belong to the user", StatusCode: int64(400)}, nil
	}

	if err != nil {
		fmt.Println("Failed to connect to database")
		return &orderpb.CancelOrderResponse{Message: "", Error: "Internal Server Error", StatusCode: int64(500)}, nil
	}

	// delete the user..
	res = orderDBConnector.Delete(&order)
	if res.Error != nil {
		fmt.Println("Failed to delete the order")
		return &orderpb.CancelOrderResponse{Message: "", Error: "Internal server error", StatusCode: int64(500)}, nil
	}

	return &orderpb.CancelOrderResponse{ Data: &orderpb.CancelOrderData{ OrderId: request.OrderId }, Message: "Order Cancelled Successfully", Error: "", StatusCode: int64(200)}, nil
}
