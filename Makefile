.PHONY: all migrate_up migrate_down

all:
	docker-compose up

migrate_up:
	goose up

migrate_down:
	goose down
