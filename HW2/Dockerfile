FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/ogen-go/ogen/cmd/ogen@v1.20.1
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

COPY . .

RUN make build

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/hw2 .

EXPOSE 8080

CMD ["./hw2"]
