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

# Accept a build argument to determine which binary to build
ARG TARGET=api

# Build the Go app
RUN go build -o /build/${TARGET} ./cmd/${TARGET}

# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Accept the same build arg to copy the right binary
ARG TARGET=main

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /build/${TARGET} .

# Expose port 8080 to the outside world
EXPOSE 8080

RUN apk add libc6-compat

# Command to run the executable
CMD ["./${TARGET}"]
