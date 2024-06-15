package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"order-microservice/config"
	"order-microservice/jwt"
	orderpb "order-microservice/proto/order"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
}

var orderDBConnector *gorm.DB
var orderItemDBConnector *gorm.DB

func startServer() {
	godotenv.Load(".env")
	fmt.Println("Starting order-microservice server...")
	// Connect to the database
	orderDB, orderItemDB, err := config.ConnectDB()
	orderDBConnector = orderDB
	orderItemDBConnector = orderItemDB
	
	
	if err != nil {
		log.Fatalf("Could not connect to the database: %s", err)
	}
	// Start the gRPC server
	listner, err := net.Listen("tcp", "localhost:50053")
	// Check if there is an error while starting the server
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	// Create a new gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(jwt.UnaryInterceptor),
	)

	// Register the service with the server
	orderpb.RegisterOrderServiceServer(grpcServer, &OrderService{})

	// Start the server in a new goroutine (concurrency) (Serve).
	go func() {
		if err := grpcServer.Serve(listner); err != nil {
			log.Fatalf("Failed to serve: %s", err)
		}
	}()
	// Create a new gRPC-Gateway server (gateway).
	connection, err := grpc.DialContext(
		context.Background(),
		"localhost:50053",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	// Create a new gRPC-Gateway mux (gateway).
	gwmux := runtime.NewServeMux()

	// Register the service with the server (gateway).
	err = orderpb.RegisterOrderServiceHandler(context.Background(), gwmux, connection)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	// Create a new HTTP server (gateway). (Serve). (ListenAndServe)
	gwServer := &http.Server{
		Addr:    ":8093",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8093")
	log.Fatalln(gwServer.ListenAndServe())
}

func main() {
	startServer()
}
