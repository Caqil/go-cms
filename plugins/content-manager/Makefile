.PHONY: build clean

PLUGIN_NAME=content-manager
OUTPUT_FILE=../$(PLUGIN_NAME).so

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@go build -buildmode=plugin -o $(OUTPUT_FILE) .
	@echo "✓ Plugin built successfully: $(OUTPUT_FILE)"

clean:
	@rm -f $(OUTPUT_FILE)
	@echo "✓ Cleaned $(PLUGIN_NAME) plugin"

test:
	@go test -v ./...

install: build
	@echo "✓ Plugin installed and ready to use"