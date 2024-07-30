# Base image
FROM golang:1.22.2-alpine3.19

# Move to working directory /app
WORKDIR /app

RUN apk add --no-cache bash


# Copy the code into the container
COPY . .

# Download dependencies using go mod
RUN go mod tidy && go mod vendor

# Expose PORT 8093 for the order microservice grpc gateway
EXPOSE 8093
# Expose PORT 8089 for payment gateway.
EXPOSE 8089
# Command to run the application when starting the container
CMD ["go", "run", "."]