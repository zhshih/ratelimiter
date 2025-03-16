# Stage 1: Build the Go application
FROM golang:1.21 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to manage dependencies
COPY go.mod go.sum ./

# Download and cache dependencies
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ratelimiter cmd/main.go

# Stage 2: Create a smaller image to run the app
FROM alpine:latest

# Install necessary dependencies for running the app (if any)
RUN apk --no-cache add ca-certificates

# Set the working directory in the new image
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/ratelimiter .

# Expose the ports the app will run on
EXPOSE 20001
EXPOSE 50001
EXPOSE 50011

# Command to run the Go application
CMD ["./ratelimiter"]
