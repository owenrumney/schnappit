.PHONY: build run clean deps test bundle

APP_NAME := schnappit
BUILD_DIR := build
CMD_DIR := cmd/schnappit

# Build the application
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

# Run tests
test:
	go test -v ./...

# Create macOS application bundle
bundle: build
	@echo "Creating application bundle..."
	@mkdir -p $(BUILD_DIR)/Schnappit.app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/Schnappit.app/Contents/Resources
	@cp $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/Schnappit.app/Contents/MacOS/
	@cp Info.plist $(BUILD_DIR)/Schnappit.app/Contents/
	@echo "Bundle created at $(BUILD_DIR)/Schnappit.app"
	@echo "Install with: cp -r $(BUILD_DIR)/Schnappit.app /Applications/"

# Development: build and run with verbose logging
dev:
	go run ./$(CMD_DIR)

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run
