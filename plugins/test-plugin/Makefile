.PHONY: build clean test zip

PLUGIN_NAME=test-plugin
OUTPUT_FILE=$(PLUGIN_NAME).so

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@go build -buildmode=plugin -o $(OUTPUT_FILE) .
	@echo "✓ Plugin built successfully: $(OUTPUT_FILE)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(OUTPUT_FILE)
	@rm -f $(PLUGIN_NAME).zip
	@echo "✓ Clean completed"

test:
	@echo "Running tests..."
	@go test ./...

zip: clean
	@echo "Creating plugin zip package..."
	@zip -r $(PLUGIN_NAME).zip . -x "*.so" "*.zip" ".git/*" ".DS_Store"
	@echo "✓ Plugin package created: $(PLUGIN_NAME).zip"

install: zip
	@echo "Installing plugin..."
	@curl -X POST \
		-F "plugin=@$(PLUGIN_NAME).zip" \
		http://localhost:8080/api/v1/admin/plugins/upload \
		-H "Authorization: Bearer YOUR_TOKEN_HERE"

dev: build
	@echo "Development build completed"

.DEFAULT_GOAL := build
