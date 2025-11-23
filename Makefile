.PHONY: all migrate_up migrate_down lint test

all:
	docker-compose up

migrate_up:
	goose up

migrate_down:
	goose down

lint:
	golangci-lint run ./...

test:
	go test -v ./...
