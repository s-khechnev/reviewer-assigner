DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= db
DB_USER ?= postgres
DB_PASSWORD ?= 12345
DATA_DIR := data

PSQL_CMD = PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

.PHONY: gen clean load clean_db all

all: clean_db gen upload load
	@echo "All tasks completed successfully!"

gen:
	cd $(DATA_DIR) && go run gen.go

clean:
	rm -f *.data.json && cd $(DATA_DIR) && rm -f *.csv

upload: $(DATA_DIR)/teams.csv $(DATA_DIR)/users.csv $(DATA_DIR)/pull_requests.csv $(DATA_DIR)/pull_request_reviewers.csv
	@echo "Uploading CSV data to database..."

	@echo "Uploading teams..."
	$(PSQL_CMD) -c "\copy teams FROM '$(DATA_DIR)/teams.csv' DELIMITER ',' CSV HEADER;"
	$(PSQL_CMD) -c "SELECT pg_get_serial_sequence('teams', 'id');"

	@echo "Uploading users..."
	$(PSQL_CMD) -c "\copy users FROM '$(DATA_DIR)/users.csv' DELIMITER ',' CSV HEADER;"

	@echo "Uploading pull_requests..."
	$(PSQL_CMD) -c "\copy pull_requests FROM '$(DATA_DIR)/pull_requests.csv' DELIMITER ',' CSV HEADER;"

	@echo "Uploading pull_request_reviewers..."
	$(PSQL_CMD) -c "\copy pull_request_reviewers FROM '$(DATA_DIR)/pull_request_reviewers.csv' DELIMITER ',' CSV HEADER;"

	@echo "Update seqs..."
	$(PSQL_CMD) -f update_seqs.sql

	@echo "Data uploaded successfully"

load:
	@echo "Loading all test data from database..."

	@echo "Loading users..."
	$(PSQL_CMD) -t -A -F"," -c "SELECT json_agg(u.user_id) FROM users u" > users.data.json

	@echo "Loading teams..."
	$(PSQL_CMD) -t -A -F"," -f load_teams.sql > teams.data.json

	@echo "Loading pull requests..."
	$(PSQL_CMD) -t -A -F"," -f load_pull_requests.sql > prs.data.json

clean_db:
	cd .. && goose reset && goose up
