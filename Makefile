.PHONY: build run test docker-build docker-up clean

BINARY_NAME=gpm

build:
	go build -o $(BINARY_NAME) cmd/gpm/main.go

run: build
	./$(BINARY_NAME) -config configs/config.yaml

test:
	go test ./...

docker-build:
	docker build -t $(BINARY_NAME):latest -f build/Dockerfile .

docker-up:
	docker-compose -f deploy/docker-compose.yml up -d

clean:
	rm -f $(BINARY_NAME)
	go clean
