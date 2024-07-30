package main

import (
	"context"
	"net"
	"net/http"
	"order-microservice/config"
	"order-microservice/jwt"
	orderpb "order-microservice/proto/order"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
}

const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusInternalServerError = 500
)

var (
	orderDBConnector     *gorm.DB
	orderItemDBConnector *gorm.DB
	logger               *zap.Logger
)

func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
}

func startServer() {
	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("Failed to load .env file", zap.Error(err))
	} else {
		logger.Info("Loaded .env file")
	}

	// Connect to the database
	orderDB, orderItemDB, err := config.ConnectDB()
	if err != nil {
		logger.Fatal("Could not connect to the database", zap.Error(err))
	}
	orderDBConnector = orderDB
	orderItemDBConnector = orderItemDB
	logger.Info("Connected to the database")
	// Start the gRPC server
	listner, err := net.Listen("tcp", "localhost:50053")
	// Check if there is an error while starting the server
	if err != nil {
		logger.Fatal("Failed to start listener", zap.Error(err))
	}
	// Create a new gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(jwt.UnaryInterceptor),
	)

	// Register the service with the server
	orderpb.RegisterOrderServiceServer(grpcServer, &OrderService{})
	logger.Info("gRPC server started on localhost:50053")
	// Start the server in a new goroutine (concurrency) (Serve).
	go func() {
		if err := grpcServer.Serve(listner); err != nil {
			logger.Fatal("Failed to serve gRPC server", zap.Error(err))
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
		logger.Fatal("Failed to dial server", zap.Error(err))
	}
	// Create a new gRPC-Gateway mux (gateway).
	gwmux := runtime.NewServeMux()

	// Register the service with the server (gateway).
	err = orderpb.RegisterOrderServiceHandler(context.Background(), gwmux, connection)
	if err != nil {
		logger.Fatal("Failed to register gRPC-Gateway", zap.Error(err))
	}
	// Enable CORS
	corsOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	corsHandler := handlers.CORS(corsOrigins, corsMethods, corsHeaders)
	wrappedGwmux := corsHandler(gwmux)

	// Create a new HTTP server (gateway). (Serve). (ListenAndServe)
	gwServer := &http.Server{
		Addr:    ":8093",
		Handler: wrappedGwmux,
	}

	logger.Info("Serving gRPC-Gateway on http://0.0.0.0:8093")
	logger.Fatal("Failed to serve gRPC-Gateway", zap.Error(gwServer.ListenAndServe()))

}

func main() {
	startServer()
}
