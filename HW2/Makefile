SPEC     := api/openapi.yaml
SQLC_CFG := sqlc.yaml

OGEN_OUT    := .build/openapi
API_PACKAGE := product_api
DB_PACKAGE  := mega_db

BINARY := hw2

.PHONY: generate build migrate clean up down logs

generate:
	ogen --target $(OGEN_OUT) --package $(API_PACKAGE) $(SPEC)
	sqlc generate -f $(SQLC_CFG)

build: generate
	go build -o $(BINARY) .

migrate:
	flyway -configFiles=flyway.conf migrate

clean:
	rm -rf $(OGEN_OUT)
	rm -f $(BINARY)

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f app
