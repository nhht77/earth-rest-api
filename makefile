.PHONY: start build stop clean db-init db-stop

DB_CONTAINER = earth-rest-api-db
DB_PORT = 5433
MIGRATIONS_DIR = $(CURDIR)/migrations

# Build the server binary
build:
	cd server && go build -o server .

# Source the env file and run the service (foreground)
# - --port 8090: host 8080 is occupied by the cm-backend dev stack
start: build
	. ./env && cd server && ./server --port 8090 --database-port 5433

# Test mode: uses TEST_* database, clears tables after run
start-test: build
	. ./env && cd server && ./server --test-build --port 8090 --database-port 5433

# Remove the local postgres and recreate it with the migrations mounted at
# /docker-entrypoint-initdb.d — the image creates the schema on first init.
# WARNING: data in the container volume is lost (fine for local dev).
db-init:
	docker rm -f $(DB_CONTAINER) 2>/dev/null || true
	docker run -d --name $(DB_CONTAINER) \
		-e POSTGRES_USER=earth -e POSTGRES_PASSWORD=earth -e POSTGRES_DB=earth \
		-p $(DB_PORT):5432 \
		-v $(MIGRATIONS_DIR):/docker-entrypoint-initdb.d:ro \
		postgres:16-alpine

# Stop the local postgres (data kept in the container, remove with db-init)
db-stop:
	docker stop $(DB_CONTAINER)

clean:
	rm -f server/server

