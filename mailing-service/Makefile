.PHONY: run build test clean

run:
	go run ./cmd/server

build:
	go build -o bin/mailing-service ./cmd/server

test:
	go test ./...

race:
	go run -race ./cmd/server

lint:
	go vet ./...

clean:
	rm -rf bin/

docker-build:
	docker compose build

docker-run:
	docker compose up -d

docker-stop:
	docker compose down
