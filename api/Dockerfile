# Start from the official Golang image as the base image
FROM golang:1.23.10 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o api ./cmd/api
RUN go build -o workers ./cmd/workers
RUN go build -o migrate ./cmd/migrate


# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/api .
COPY --from=builder /app/workers .
COPY --from=builder /app/migrate .


# Expose port 8080 to the outside world
EXPOSE 8080

RUN apk add --no-cache libc6-compat tzdata
