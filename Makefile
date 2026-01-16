.PHONY: build run clean deps test bundle icon

APP_NAME := schnappit
BUILD_DIR := build
CMD_DIR := cmd/schnappit
ICONGEN_DIR := cmd/icongen

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

# Generate app icon
icon:
	@echo "Generating app icon..."
	@go build -o $(BUILD_DIR)/icongen ./$(ICONGEN_DIR)
	@$(BUILD_DIR)/icongen $(BUILD_DIR)/icon_tmp
	@mkdir -p $(BUILD_DIR)/AppIcon.iconset
	@cp $(BUILD_DIR)/icon_tmp/icon_16x16.png $(BUILD_DIR)/AppIcon.iconset/icon_16x16.png
	@cp $(BUILD_DIR)/icon_tmp/icon_32x32.png $(BUILD_DIR)/AppIcon.iconset/icon_16x16@2x.png
	@cp $(BUILD_DIR)/icon_tmp/icon_32x32.png $(BUILD_DIR)/AppIcon.iconset/icon_32x32.png
	@cp $(BUILD_DIR)/icon_tmp/icon_64x64.png $(BUILD_DIR)/AppIcon.iconset/icon_32x32@2x.png
	@cp $(BUILD_DIR)/icon_tmp/icon_128x128.png $(BUILD_DIR)/AppIcon.iconset/icon_128x128.png
	@cp $(BUILD_DIR)/icon_tmp/icon_256x256.png $(BUILD_DIR)/AppIcon.iconset/icon_128x128@2x.png
	@cp $(BUILD_DIR)/icon_tmp/icon_256x256.png $(BUILD_DIR)/AppIcon.iconset/icon_256x256.png
	@cp $(BUILD_DIR)/icon_tmp/icon_512x512.png $(BUILD_DIR)/AppIcon.iconset/icon_256x256@2x.png
	@cp $(BUILD_DIR)/icon_tmp/icon_512x512.png $(BUILD_DIR)/AppIcon.iconset/icon_512x512.png
	@cp $(BUILD_DIR)/icon_tmp/icon_1024x1024.png $(BUILD_DIR)/AppIcon.iconset/icon_512x512@2x.png
	@iconutil -c icns $(BUILD_DIR)/AppIcon.iconset -o $(BUILD_DIR)/AppIcon.icns
	@rm -rf $(BUILD_DIR)/icon_tmp $(BUILD_DIR)/AppIcon.iconset $(BUILD_DIR)/icongen
	@echo "Icon generated at $(BUILD_DIR)/AppIcon.icns"

# Create macOS application bundle
bundle: build icon
	@echo "Creating application bundle..."
	@mkdir -p $(BUILD_DIR)/Schnappit.app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/Schnappit.app/Contents/Resources
	@cp $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/Schnappit.app/Contents/MacOS/
	@cp $(BUILD_DIR)/AppIcon.icns $(BUILD_DIR)/Schnappit.app/Contents/Resources/
	@cp Info.plist $(BUILD_DIR)/Schnappit.app/Contents/
	@codesign --force --deep --sign - $(BUILD_DIR)/Schnappit.app || true
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
