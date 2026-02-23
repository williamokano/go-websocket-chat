.PHONY: build-server build-tui build run-server run-tui docker-up docker-down docker-logs

build-server:
	go build -o bin/server ./cmd/server

build-tui:
	go build -o bin/tui ./cmd/tui

build: build-server build-tui

run-server:
	go run ./cmd/server

run-tui:
	go run ./cmd/tui -- --server ws://localhost:8080

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
