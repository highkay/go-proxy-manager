.PHONY: build run test docker-build docker-up docker-down clean

BINARY_NAME=gpm

build:
	go build -o $(BINARY_NAME) cmd/gpm/main.go

run: build
	./$(BINARY_NAME) -config configs/config.yaml

test:
	go test ./...

docker-build:
	docker compose -f deploy/docker-compose.yml build --no-cache

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down

clean:
	rm -f $(BINARY_NAME)
	go clean
