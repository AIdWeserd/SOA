.PHONY: generate up down logs build

generate:
	@echo "Generating protobuf code for booking-service..."
	protoc \
		--proto_path=proto \
		--go_out=booking-service/gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=booking-service/gen \
		--go-grpc_opt=paths=source_relative \
		flight/v1/flight.proto
	@echo "Generating protobuf code for flight-service..."
	protoc \
		--proto_path=proto \
		--go_out=flight-service/gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=flight-service/gen \
		--go-grpc_opt=paths=source_relative \
		flight/v1/flight.proto

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

build:
	docker compose build
