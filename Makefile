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
	@cp Info.plist $(BUILD_DIR)/Schnappit.app/Contents/ 2>/dev/null || \
		echo '<?xml version="1.0" encoding="UTF-8"?>\n\
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n\
<plist version="1.0">\n\
<dict>\n\
	<key>CFBundleExecutable</key>\n\
	<string>schnappit</string>\n\
	<key>CFBundleIdentifier</key>\n\
	<string>com.owenrumney.schnappit</string>\n\
	<key>CFBundleName</key>\n\
	<string>Schnappit</string>\n\
	<key>CFBundleVersion</key>\n\
	<string>1.0.0</string>\n\
	<key>CFBundleShortVersionString</key>\n\
	<string>1.0.0</string>\n\
	<key>LSMinimumSystemVersion</key>\n\
	<string>12.3</string>\n\
	<key>NSHighResolutionCapable</key>\n\
	<true/>\n\
	<key>NSScreenCaptureUsageDescription</key>\n\
	<string>Schnappit needs screen recording permission to capture screenshots.</string>\n\
	<key>NSAppleEventsUsageDescription</key>\n\
	<string>Schnappit needs accessibility permission for global hotkeys.</string>\n\
</dict>\n\
</plist>' > $(BUILD_DIR)/Schnappit.app/Contents/Info.plist
	@echo "Bundle created at $(BUILD_DIR)/Schnappit.app"

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
