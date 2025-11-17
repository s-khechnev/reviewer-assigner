.PHONY: migrate_up migrate_down

migrate_up:
	goose up

migrate_down:
	goose down
