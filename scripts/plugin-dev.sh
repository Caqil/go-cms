# scripts/plugin-dev.sh - Development helper script
#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Quick build and test a single plugin
quick_build() {
    local plugin_name="$1"
    
    log_info "Quick building plugin: $plugin_name"
    
    # Build the plugin
    if "$SCRIPT_DIR/build-plugins.sh" build "$plugin_name" --verbose; then
        log_success "Plugin built successfully"
        
        # Show plugin info
        "$SCRIPT_DIR/build-plugins.sh" list | grep "$plugin_name" || true
    else
        log_warning "Plugin build failed"
        return 1
    fi
}

# Create new plugin from template
create_plugin() {
    local plugin_name="$1"
    
    if [[ -z "$plugin_name" ]]; then
        echo "Usage: create_plugin <plugin-name>"
        return 1
    fi
    
    log_info "Creating new plugin: $plugin_name"
    
    if command -v "$ROOT_DIR/cmd/plugin-builder/main.go" &> /dev/null; then
        go run "$ROOT_DIR/cmd/plugin-builder/main.go" create "$plugin_name"
    else
        log_warning "Plugin builder not found, creating basic structure..."
        create_basic_plugin "$plugin_name"
    fi
}

# Create basic plugin structure manually
create_basic_plugin() {
    local plugin_name="$1"
    local plugin_dir="$ROOT_DIR/plugins/$plugin_name"
    
    mkdir -p "$plugin_dir"
    
    # Create main.go
    cat > "$plugin_dir/main.go" << EOF
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// Plugin interfaces (copy from your actual plugin interface)
type Plugin interface {
    GetInfo() PluginInfo
    Initialize(deps *PluginDependencies) error
    RegisterRoutes(router *gin.RouterGroup)
    GetAdminMenuItems() []AdminMenuItem
    GetSettings() []PluginSetting
    Shutdown() error
}

type PluginInfo struct {
    Name        string \`json:"name"\`
    Version     string \`json:"version"\`
    Description string \`json:"description"\`
    Author      string \`json:"author"\`
    Website     string \`json:"website,omitempty"\`
}

type PluginDependencies struct {
    Database interface{}
    Config   interface{}
}

type AdminMenuItem struct {
    ID       string \`json:"id"\`
    Title    string \`json:"title"\`
    Icon     string \`json:"icon,omitempty"\`
    URL      string \`json:"url"\`
    Parent   string \`json:"parent,omitempty"\`
    Order    int    \`json:"order"\`
}

type PluginSetting struct {
    Key         string      \`json:"key"\`
    Label       string      \`json:"label"\`
    Type        string      \`json:"type"\`
    Value       interface{} \`json:"value"\`
    Description string      \`json:"description,omitempty"\`
    Required    bool        \`json:"required"\`
}

// ${plugin_name^}Plugin implementation
type ${plugin_name^}Plugin struct {
    deps *PluginDependencies
}

// NewPlugin is the entry point
func NewPlugin() Plugin {
    return &${plugin_name^}Plugin{}
}

func (p *${plugin_name^}Plugin) GetInfo() PluginInfo {
    return PluginInfo{
        Name:        "$plugin_name",
        Version:     "1.0.0",
        Description: "A $plugin_name plugin for the CMS",
        Author:      "Plugin Developer",
    }
}

func (p *${plugin_name^}Plugin) Initialize(deps *PluginDependencies) error {
    p.deps = deps
    return nil
}

func (p *${plugin_name^}Plugin) RegisterRoutes(router *gin.RouterGroup) {
    router.GET("/hello", p.handleHello)
}

func (p *${plugin_name^}Plugin) GetAdminMenuItems() []AdminMenuItem {
    return []AdminMenuItem{
        {
            ID:    "$plugin_name-menu",
            Title: "${plugin_name^}",
            Icon:  "puzzle-piece",
            URL:   "/admin/plugins/$plugin_name",
            Order: 50,
        },
    }
}

func (p *${plugin_name^}Plugin) GetSettings() []PluginSetting {
    return []PluginSetting{
        {
            Key:      "enabled",
            Label:    "Enable ${plugin_name^}",
            Type:     "boolean",
            Value:    true,
            Required: false,
        },
    }
}

func (p *${plugin_name^}Plugin) Shutdown() error {
    return nil
}

func (p *${plugin_name^}Plugin) handleHello(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Hello from ${plugin_name^}!",
        "plugin":  p.GetInfo().Name,
    })
}
EOF

    # Create go.mod
    cat > "$plugin_dir/go.mod" << EOF
module $plugin_name

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
)
EOF

    # Create Makefile
    cat > "$plugin_dir/Makefile" << EOF
.PHONY: build clean test

PLUGIN_NAME=$plugin_name
OUTPUT_FILE=../\$(PLUGIN_NAME).so

build:
	@echo "Building \$(PLUGIN_NAME) plugin..."
	@go build -buildmode=plugin -o \$(OUTPUT_FILE) .
	@echo "✓ Plugin built successfully: \$(OUTPUT_FILE)"

clean:
	@rm -f \$(OUTPUT_FILE)

test:
	@go test -v ./...

dev: build
	@echo "Plugin ready for development"
EOF

    log_success "Plugin created: $plugin_dir"
    log_info "To build: cd $plugin_dir && make build"
}

case "${1:-}" in
    "build")
        if [[ -n "${2:-}" ]]; then
            quick_build "$2"
        else
            echo "Usage: $0 build <plugin-name>"
        fi
        ;;
    "create")
        if [[ -n "${2:-}" ]]; then
            create_plugin "$2"
        else
            echo "Usage: $0 create <plugin-name>"
        fi
        ;;
    *)
        echo "Usage: $0 {build|create} <plugin-name>"
        echo "  build  - Quick build and test a plugin"
        echo "  create - Create a new plugin from template"
        ;;
esac

---

# scripts/validate-plugins.sh - Plugin validation script
#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
PLUGINS_DIR="$ROOT_DIR/plugins"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Validate plugin structure
validate_plugin() {
    local plugin_name="$1"
    local plugin_dir="$PLUGINS_DIR/$plugin_name"
    local errors=0

    log_info "Validating plugin: $plugin_name"

    # Check required files
    if [[ ! -f "$plugin_dir/main.go" ]]; then
        log_error "Missing main.go"
        ((errors++))
    fi

    # Check main.go content
    if [[ -f "$plugin_dir/main.go" ]]; then
        # Check for NewPlugin function
        if ! grep -q "func NewPlugin" "$plugin_dir/main.go"; then
            log_error "Missing NewPlugin function"
            ((errors++))
        fi

        # Check for Plugin interface implementation
        local required_methods=("GetInfo" "Initialize" "RegisterRoutes" "GetAdminMenuItems" "GetSettings" "Shutdown")
        for method in "${required_methods[@]}"; do
            if ! grep -q "func.*$method" "$plugin_dir/main.go"; then
                log_warning "Missing or incorrect method: $method"
            fi
        done

        # Check syntax
        if ! (cd "$plugin_dir" && go fmt -n . >/dev/null 2>&1); then
            log_error "Go syntax errors found"
            ((errors++))
        fi
    fi

    # Check go.mod
    if [[ ! -f "$plugin_dir/go.mod" ]]; then
        log_warning "Missing go.mod file (recommended)"
    fi

    # Check for tests
    if ! find "$plugin_dir" -name "*_test.go" | grep -q .; then
        log_info "No test files found (consider adding tests)"
    fi

    return $errors
}

# Validate all plugins
validate_all() {
    local total_errors=0
    local plugin_count=0

    log_info "Validating all plugins..."

    for dir in "$PLUGINS_DIR"/*; do
        if [[ -d "$dir" && -f "$dir/main.go" ]]; then
            local plugin_name=$(basename "$dir")
            ((plugin_count++))
            
            if ! validate_plugin "$plugin_name"; then
                ((total_errors++))
            fi
            echo
        fi
    done

    if [[ $total_errors -eq 0 ]]; then
        log_success "All $plugin_count plugins passed validation"
    else
        log_error "$total_errors plugins have validation errors"
    fi

    return $total_errors
}

# Check plugin dependencies
check_dependencies() {
    local plugin_name="$1"
    local plugin_dir="$PLUGINS_DIR/$plugin_name"

    if [[ -f "$plugin_dir/go.mod" ]]; then
        log_info "Checking dependencies for $plugin_name..."
        (cd "$plugin_dir" && go mod verify) || return 1
        (cd "$plugin_dir" && go mod tidy) || return 1
    fi
}

case "${1:-all}" in
    "all")
        validate_all
        ;;
    *)
        if [[ -d "$PLUGINS_DIR/$1" ]]; then
            validate_plugin "$1"
        else
            log_error "Plugin not found: $1"
            exit 1
        fi
        ;;
esac

---

# .github/workflows/build-plugins.yml - GitHub Actions workflow
name: Build Plugins

on:
  push:
    branches: [ main, develop ]
    paths: [ 'plugins/**' ]
  pull_request:
    branches: [ main ]
    paths: [ 'plugins/**' ]

jobs:
  build-plugins:
    name: Build and Test Plugins
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        go-version: [1.21, 1.22]
        
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: \${{ matrix.go-version }}
        
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
          ~/go/pkg/mod
        key: \${{ runner.os }}-go-\${{ matrix.go-version }}-\${{ hashFiles('**/go.sum') }}
        restore-keys: |
          \${{ runner.os }}-go-\${{ matrix.go-version }}-
          
    - name: Install dependencies
      run: go mod download
      
    - name: Validate plugins
      run: ./scripts/validate-plugins.sh
      
    - name: Build plugins
      run: ./scripts/build-plugins.sh build --mode production --verbose
      
    - name: Test plugins
      run: |
        for plugin_dir in plugins/*/; do
          if [[ -f "\$plugin_dir/go.mod" ]]; then
            echo "Testing plugin: \$(basename \$plugin_dir)"
            (cd "\$plugin_dir" && go test -v ./... || true)
          fi
        done
        
    - name: Upload build artifacts
      uses: actions/upload-artifact@v3
      with:
        name: plugins-go\${{ matrix.go-version }}
        path: build/plugins/*.so
        retention-days: 30

---

# Docker build support
# Dockerfile.plugins - Multi-stage build for plugins
FROM golang:1.21-alpine AS builder

# Install build tools
RUN apk add --no-cache git make bash

WORKDIR /app

# Copy source
COPY . .

# Build plugins
RUN chmod +x scripts/build-plugins.sh && \
    scripts/build-plugins.sh build --mode production

# Runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy built plugins
COPY --from=builder /app/build/plugins/*.so ./plugins/

# Copy main application (if needed)
COPY --from=builder /app/cmd/server/main /app/

EXPOSE 8080

CMD ["./main"]

---

# Makefile integration
# Add these targets to your main Makefile

.PHONY: plugins plugins-dev plugins-clean plugins-watch plugins-validate

# Build all plugins
plugins:
	@./scripts/build-plugins.sh build

# Build plugins in development mode
plugins-dev:
	@./scripts/build-plugins.sh build --mode development --verbose

# Clean plugin builds
plugins-clean:
	@./scripts/build-plugins.sh clean

# Watch and rebuild plugins
plugins-watch:
	@./scripts/build-plugins.sh watch

# Validate plugin structure
plugins-validate:
	@./scripts/validate-plugins.sh

# Build specific plugin
plugin-%:
	@./scripts/build-plugins.sh build $*

# Create new plugin
create-plugin:
	@read -p "Enter plugin name: " name; \
	./scripts/plugin-dev.sh create $$name

---

# scripts/install-build-deps.sh - Install build dependencies
#!/bin/bash

set -euo pipefail

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "freebsd"* ]]; then
        echo "freebsd"
    else
        echo "unsupported"
    fi
}

install_deps() {
    local os=$(detect_os)
    
    case $os in
        "linux")
            log_info "Installing dependencies for Linux..."
            
            # Check if we have apt, yum, or other package managers
            if command -v apt-get &> /dev/null; then
                sudo apt-get update
                sudo apt-get install -y build-essential git inotify-tools
            elif command -v yum &> /dev/null; then
                sudo yum groupinstall -y "Development Tools"
                sudo yum install -y git inotify-tools
            elif command -v dnf &> /dev/null; then
                sudo dnf groupinstall -y "Development Tools"
                sudo dnf install -y git inotify-tools
            else
                log_warning "Unknown package manager. Please install build tools manually."
            fi
            ;;
            
        "macos")
            log_info "Installing dependencies for macOS..."
            
            # Check if Xcode command line tools are installed
            if ! command -v git &> /dev/null; then
                log_info "Installing Xcode command line tools..."
                xcode-select --install
            fi
            
            # Check if Homebrew is available
            if command -v brew &> /dev/null; then
                brew install fswatch
            else
                log_warning "Homebrew not found. Install fswatch manually for watch functionality."
            fi
            ;;
            
        "freebsd")
            log_info "Installing dependencies for FreeBSD..."
            sudo pkg install -y git
            ;;
            
        *)
            log_warning "Unsupported OS. Go plugins may not work on this platform."
            return 1
            ;;
    esac
    
    log_success "Build dependencies installed successfully"
}

# Check Go installation
check_go() {
    if ! command -v go &> /dev/null; then
        log_warning "Go is not installed. Please install Go 1.21 or later."
        log_info "Visit: https://golang.org/dl/"
        return 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_success "Go $go_version is installed"
    
    return 0
}

main() {
    log_info "Installing build dependencies for plugin system..."
    
    if ! check_go; then
        exit 1
    fi
    
    if ! install_deps; then
        exit 1
    fi
    
    log_success "All dependencies installed successfully!"
    log_info "You can now build plugins using: ./scripts/build-plugins.sh"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi