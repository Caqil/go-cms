.PHONY: build clean test

PLUGIN_NAME=hello-world
OUTPUT_FILE=../$(PLUGIN_NAME).so

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@go build -buildmode=plugin -o $(OUTPUT_FILE) .
	@echo "Plugin built successfully: $(OUTPUT_FILE)"

clean:
	@echo "Cleaning up..."
	@rm -f $(OUTPUT_FILE)
	@go clean

test:
	@echo "Running tests..."
	@go test -v ./...

install: build
	@echo "Installing plugin..."
	@cp $(OUTPUT_FILE) ../../plugins/

dev: clean build
	@echo "Development build complete"

.DEFAULT_GOAL := build