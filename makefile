.PHONY: start build stop clean

# Build the server binary
build:
	cd server && go build -o server .

# Source the env file and run the service (foreground)
# --database-port 5433: the earth-rest-api-db container publishes 5433 (5432 is taken
# by the small-backend postgres)
start: build
	. ./env && ./server/server --database-port 5433

# Test mode: uses TEST_* database, clears tables after run
start-test: build
	. ./env && ./server/server --test-build --database-port 5433

clean:
	rm -f server/server

