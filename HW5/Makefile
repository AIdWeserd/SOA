.PHONY: up down logs logs-movie logs-aggregation ps test clean

up:
	docker compose up --build -d

down:
	docker compose down -v

logs:
	docker compose logs -f

logs-movie:
	docker compose logs -f movie-service

logs-aggregation:
	docker compose logs -f aggregation-service

ps:
	docker compose ps

test:
	pip install -q -r tests/requirements.txt
	pytest tests/test_pipeline.py -v

clean:
	docker compose down -v --rmi local
	docker volume prune -f
