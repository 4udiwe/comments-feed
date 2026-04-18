.PHONY: up-memory up-postgres down

up-memory:
	STORAGE_TYPE=memory docker compose up --build -d

up-postgres:
	STORAGE_TYPE=postgres docker compose --profile postgres up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

test:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out