# Makefile for go-parallel-var-cmds
# Build backend and frontend, run tests, and start the server.

# Backend binary output directory
BIN_DIR := bin

.PHONY: build-backend build-frontend test serve clean

# Build the Go REST server
build-backend:
	@echo "Building Go backend..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/server ./cmd/server

# Build the React frontend using npm and vite
build-frontend:
	@echo "Building React frontend..."
	@cd frontend && npm install && npm run build

# Run Go unit tests
# This will run all tests in the repository
# including database and executor packages
# and report coverage statistics
# The CI workflow will call this target
# to ensure tests pass on every commit
 test:
	@echo "Running backend tests..."
	@go test ./...

# Start the development server
# This runs the Go server. In a real development
# workflow you might run `npm start` separately
# for the frontend, but for simplicity we just
# start the Go server here. The frontend
# should be built separately via build-frontend
serve:
	@echo "Starting Go server..."
	@go run ./cmd/server/main.go

# Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
