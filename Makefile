docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

test-unit:
	go test -short -v ./internal/service/...

test-integration:
	go test -v ./internal/repository/... ./internal/workers/...

test:
	go test -v ./...
