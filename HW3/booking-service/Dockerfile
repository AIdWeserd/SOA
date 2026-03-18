# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

# Install protoc and build tools
RUN apk add --no-cache protobuf-dev git

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

WORKDIR /app

# Copy go module files
COPY booking-service/go.mod ./

# Copy proto definitions
COPY proto/ /proto/

# Copy service source (gen/ only has .gitkeep at this point)
COPY booking-service/ .

# Generate protobuf code (overwrites the empty gen/ placeholder)
RUN mkdir -p gen/flight/v1 && \
    protoc \
      --proto_path=/proto \
      --proto_path=/usr/include \
      --go_out=gen \
      --go_opt=paths=source_relative \
      --go-grpc_out=gen \
      --go-grpc_opt=paths=source_relative \
      flight/v1/flight.proto

# Download deps and tidy (go.sum will be regenerated)
RUN go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /booking-service ./cmd/main.go

# ---
FROM alpine:3.19 AS final

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /booking-service /app/booking-service
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/booking-service"]
