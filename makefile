.PHONY: start build stop clean

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

clean:
	rm -f server/server

