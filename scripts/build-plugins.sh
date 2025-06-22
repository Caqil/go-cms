#!/bin/bash

# build-plugins.sh - Comprehensive plugin build script for Go CMS
# Usage: ./scripts/build-plugins.sh [command] [options]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
PLUGINS_DIR="$ROOT_DIR/plugins"
BUILD_DIR="$ROOT_DIR/build/plugins"
LOG_DIR="$ROOT_DIR/logs"
LOG_FILE="$LOG_DIR/plugin-build.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build configuration
BUILD_MODE="development"
PARALLEL_JOBS=4
VERBOSE=false
DRY_RUN=false
FORCE_BUILD=false
CLEAN_FIRST=false

# Plugin build flags
GO_BUILD_FLAGS="-buildmode=plugin"
DEVELOPMENT_FLAGS="-gcflags='all=-N -l'"
PRODUCTION_FLAGS="-ldflags='-s -w'"

# Initialize logging
init_logging() {
    mkdir -p "$LOG_DIR"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - Starting plugin build" > "$LOG_FILE"
}

# Logging functions
log_info() {
    local message="$1"
    echo -e "${BLUE}[INFO]${NC} $message"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - INFO: $message" >> "$LOG_FILE"
}

log_success() {
    local message="$1"
    echo -e "${GREEN}[SUCCESS]${NC} $message"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - SUCCESS: $message" >> "$LOG_FILE"
}

log_warning() {
    local message="$1"
    echo -e "${YELLOW}[WARNING]${NC} $message"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - WARNING: $message" >> "$LOG_FILE"
}

log_error() {
    local message="$1"
    echo -e "${RED}[ERROR]${NC} $message" >&2
    echo "$(date '+%Y-%m-%d %H:%M:%S') - ERROR: $message" >> "$LOG_FILE"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check Go version
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $go_version"

    # Check if we're on a supported platform for plugins
    local goos=$(go env GOOS)
    local goarch=$(go env GOARCH)
    
    if [[ "$goos" != "linux" && "$goos" != "darwin" && "$goos" != "freebsd" ]]; then
        log_error "Go plugins are not supported on $goos/$goarch"
        log_error "Supported platforms: linux, darwin (macOS), freebsd"
        exit 1
    fi

    # Check if plugins directory exists
    if [[ ! -d "$PLUGINS_DIR" ]]; then
        log_error "Plugins directory not found: $PLUGINS_DIR"
        exit 1
    fi

    # Create build directory
    mkdir -p "$BUILD_DIR"

    log_success "Prerequisites check passed"
}

# Discover plugin directories
discover_plugins() {
    local plugins=()
    
    log_info "Discovering plugins in $PLUGINS_DIR..."
    
    for dir in "$PLUGINS_DIR"/*; do
        if [[ -d "$dir" && -f "$dir/main.go" ]]; then
            local plugin_name=$(basename "$dir")
            # Skip if it's a .so file or other non-directory
            if [[ ! "$plugin_name" =~ \.so$ ]]; then
                plugins+=("$plugin_name")
                log_info "Found plugin: $plugin_name"
            fi
        fi
    done
    
    if [[ ${#plugins[@]} -eq 0 ]]; then
        log_warning "No plugins found in $PLUGINS_DIR"
        return 1
    fi
    
    printf '%s\n' "${plugins[@]}"
}

# Validate plugin structure
validate_plugin() {
    local plugin_name="$1"
    local plugin_dir="$PLUGINS_DIR/$plugin_name"
    
    if $VERBOSE; then
        log_info "Validating plugin: $plugin_name"
    fi
    
    # Check for required files
    if [[ ! -f "$plugin_dir/main.go" ]]; then
        log_error "Plugin $plugin_name missing main.go"
        return 1
    fi
    
    # Check if main.go has NewPlugin function
    if ! grep -q "func NewPlugin" "$plugin_dir/main.go"; then
        log_error "Plugin $plugin_name missing NewPlugin function"
        return 1
    fi
    
    # Check for go.mod (optional but recommended)
    if [[ ! -f "$plugin_dir/go.mod" ]]; then
        log_warning "Plugin $plugin_name missing go.mod file"
    fi
    
    return 0
}

# Build a single plugin
build_plugin() {
    local plugin_name="$1"
    local plugin_dir="$PLUGINS_DIR/$plugin_name"
    local output_file="$BUILD_DIR/$plugin_name.so"
    local temp_output="$BUILD_DIR/${plugin_name}.tmp.so"
    
    log_info "Building plugin: $plugin_name"
    
    # Validate plugin structure
    if ! validate_plugin "$plugin_name"; then
        log_error "Plugin validation failed: $plugin_name"
        return 1
    fi
    
    # Check if build is needed (unless force build)
    if [[ ! $FORCE_BUILD == true && -f "$output_file" ]]; then
        # Check if source is newer than output
        local source_newer=false
        while IFS= read -r -d '' file; do
            if [[ "$file" -nt "$output_file" ]]; then
                source_newer=true
                break
            fi
        done < <(find "$plugin_dir" -name "*.go" -print0)
        
        if [[ $source_newer == false ]]; then
            log_info "Plugin $plugin_name is up to date, skipping"
            return 0
        fi
    fi
    
    if $DRY_RUN; then
        log_info "[DRY RUN] Would build plugin: $plugin_name"
        return 0
    fi
    
    # Prepare build command
    local build_cmd="go build $GO_BUILD_FLAGS"
    
    # Add mode-specific flags
    case "$BUILD_MODE" in
        "development")
            build_cmd="$build_cmd $DEVELOPMENT_FLAGS"
            ;;
        "production")
            build_cmd="$build_cmd $PRODUCTION_FLAGS"
            ;;
    esac
    
    # Add output file
    build_cmd="$build_cmd -o $temp_output"
    
    # Change to plugin directory and build
    local build_start=$(date +%s)
    
    if $VERBOSE; then
        log_info "Executing: $build_cmd"
        log_info "Working directory: $plugin_dir"
    fi
    
    local build_output
    local build_result=0
    
    # Capture build output
    build_output=$(cd "$plugin_dir" && eval "$build_cmd" 2>&1) || build_result=$?
    
    local build_end=$(date +%s)
    local build_time=$((build_end - build_start))
    
    if [[ $build_result -eq 0 ]]; then
        # Move temp file to final location atomically
        mv "$temp_output" "$output_file"
        
        # Get file size
        local file_size=$(du -h "$output_file" | cut -f1)
        
        log_success "Plugin $plugin_name built successfully (${file_size}, ${build_time}s)"
        
        if $VERBOSE && [[ -n "$build_output" ]]; then
            echo "$build_output"
        fi
        
        return 0
    else
        log_error "Failed to build plugin: $plugin_name"
        log_error "Build output:"
        echo "$build_output"
        
        # Clean up temp file if it exists
        [[ -f "$temp_output" ]] && rm -f "$temp_output"
        
        return 1
    fi
}

# Build multiple plugins in parallel
build_plugins_parallel() {
    local plugins=("$@")
    local pids=()
    local results=()
    local total=${#plugins[@]}
    local success_count=0
    local failed_plugins=()
    
    log_info "Building $total plugins with $PARALLEL_JOBS parallel jobs..."
    
    # Function to build plugin in background
    build_plugin_bg() {
        local plugin="$1"
        local index="$2"
        local result_file="$BUILD_DIR/.build_result_$index"
        
        if build_plugin "$plugin"; then
            echo "SUCCESS" > "$result_file"
        else
            echo "FAILED" > "$result_file"
        fi
    }
    
    # Start builds in parallel batches
    local batch_start=0
    
    while [[ $batch_start -lt $total ]]; do
        local batch_end=$((batch_start + PARALLEL_JOBS))
        [[ $batch_end -gt $total ]] && batch_end=$total
        
        # Start batch
        local batch_pids=()
        for ((i=batch_start; i<batch_end; i++)); do
            local plugin="${plugins[$i]}"
            build_plugin_bg "$plugin" "$i" &
            batch_pids+=($!)
        done
        
        # Wait for batch to complete
        for pid in "${batch_pids[@]}"; do
            wait "$pid"
        done
        
        # Check results
        for ((i=batch_start; i<batch_end; i++)); do
            local plugin="${plugins[$i]}"
            local result_file="$BUILD_DIR/.build_result_$i"
            
            if [[ -f "$result_file" ]]; then
                local result=$(cat "$result_file")
                rm -f "$result_file"
                
                if [[ "$result" == "SUCCESS" ]]; then
                    ((success_count++))
                else
                    failed_plugins+=("$plugin")
                fi
            else
                failed_plugins+=("$plugin")
            fi
        done
        
        batch_start=$batch_end
    done
    
    # Report results
    log_info "Build completed: $success_count/$total plugins built successfully"
    
    if [[ ${#failed_plugins[@]} -gt 0 ]]; then
        log_error "Failed plugins: ${failed_plugins[*]}"
        return 1
    fi
    
    return 0
}

# Build plugins sequentially
build_plugins_sequential() {
    local plugins=("$@")
    local total=${#plugins[@]}
    local success_count=0
    local failed_plugins=()
    
    log_info "Building $total plugins sequentially..."
    
    for plugin in "${plugins[@]}"; do
        if build_plugin "$plugin"; then
            ((success_count++))
        else
            failed_plugins+=("$plugin")
        fi
    done
    
    log_info "Build completed: $success_count/$total plugins built successfully"
    
    if [[ ${#failed_plugins[@]} -gt 0 ]]; then
        log_error "Failed plugins: ${failed_plugins[*]}"
        return 1
    fi
    
    return 0
}

# Clean build artifacts
clean_plugins() {
    log_info "Cleaning plugin build artifacts..."
    
    if $DRY_RUN; then
        log_info "[DRY RUN] Would remove: $BUILD_DIR/*.so"
        return 0
    fi
    
    local cleaned_count=0
    
    # Remove .so files from build directory
    if [[ -d "$BUILD_DIR" ]]; then
        for so_file in "$BUILD_DIR"/*.so; do
            if [[ -f "$so_file" ]]; then
                rm -f "$so_file"
                ((cleaned_count++))
                log_info "Removed: $(basename "$so_file")"
            fi
        done
    fi
    
    # Remove temp files
    rm -f "$BUILD_DIR"/.build_result_*
    
    log_success "Cleaned $cleaned_count plugin files"
}

# List built plugins
list_plugins() {
    log_info "Listing plugins..."
    
    printf "%-20s %-10s %-12s %-20s\n" "PLUGIN" "STATUS" "SIZE" "MODIFIED"
    printf "%-20s %-10s %-12s %-20s\n" "------" "------" "----" "--------"
    
    local plugin_count=0
    local built_count=0
    
    if [[ -d "$PLUGINS_DIR" ]]; then
        for dir in "$PLUGINS_DIR"/*; do
            if [[ -d "$dir" && -f "$dir/main.go" ]]; then
                local plugin_name=$(basename "$dir")
                local so_file="$BUILD_DIR/$plugin_name.so"
                local status="NOT BUILT"
                local size="-"
                local modified="-"
                
                ((plugin_count++))
                
                if [[ -f "$so_file" ]]; then
                    status="BUILT"
                    size=$(du -h "$so_file" | cut -f1)
                    modified=$(date -r "$so_file" "+%Y-%m-%d %H:%M")
                    ((built_count++))
                fi
                
                printf "%-20s %-10s %-12s %-20s\n" "$plugin_name" "$status" "$size" "$modified"
            fi
        done
    fi
    
    echo
    log_info "Total plugins: $plugin_count, Built: $built_count"
}

# Watch for changes and rebuild
watch_plugins() {
    log_info "Watching plugins for changes..."
    
    if ! command -v inotifywait &> /dev/null; then
        log_error "inotifywait not found. Install inotify-tools to use watch mode"
        exit 1
    fi
    
    local plugins
    mapfile -t plugins < <(discover_plugins)
    
    log_info "Watching ${#plugins[@]} plugins for changes..."
    
    while true; do
        # Watch for changes in any plugin directory
        local changed_file
        changed_file=$(inotifywait -r -e modify,create,delete --format '%w%f' "$PLUGINS_DIR" 2>/dev/null)
        
        if [[ "$changed_file" =~ \.go$ ]]; then
            local plugin_dir=$(dirname "$changed_file")
            local plugin_name=$(basename "$plugin_dir")
            
            log_info "Change detected in $plugin_name, rebuilding..."
            
            if build_plugin "$plugin_name"; then
                log_success "Plugin $plugin_name rebuilt successfully"
            else
                log_error "Failed to rebuild plugin $plugin_name"
            fi
        fi
        
        sleep 1
    done
}

# Show help
show_help() {
    cat << EOF
build-plugins.sh - Go CMS Plugin Build Script

USAGE:
    ./scripts/build-plugins.sh [COMMAND] [OPTIONS]

COMMANDS:
    build [PLUGIN]      Build specific plugin or all plugins
    clean               Clean build artifacts
    list                List all plugins and their status
    watch               Watch for changes and auto-rebuild
    help                Show this help message

OPTIONS:
    -m, --mode MODE     Build mode: development, production (default: development)
    -j, --jobs N        Number of parallel jobs (default: 4)
    -v, --verbose       Verbose output
    -n, --dry-run       Show what would be done without executing
    -f, --force         Force rebuild even if up to date
    -c, --clean-first   Clean before building
    --sequential        Build plugins sequentially instead of parallel

EXAMPLES:
    # Build all plugins
    ./scripts/build-plugins.sh build

    # Build specific plugin
    ./scripts/build-plugins.sh build content-manager

    # Production build with verbose output
    ./scripts/build-plugins.sh build --mode production --verbose

    # Clean and rebuild all
    ./scripts/build-plugins.sh build --clean-first

    # Watch for changes
    ./scripts/build-plugins.sh watch

BUILD MODES:
    development    Fast builds with debug info (default)
    production     Optimized builds with stripped symbols

NOTES:
    - Go plugins are only supported on Linux, macOS, and FreeBSD
    - Built plugins are saved to: $BUILD_DIR
    - Build logs are saved to: $LOG_FILE

EOF
}

# Parse command line arguments
parse_args() {
    local command=""
    local plugin_name=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            build|clean|list|watch|help)
                command="$1"
                shift
                # Check if next argument is a plugin name (for build command)
                if [[ "$command" == "build" && $# -gt 0 && ! "$1" =~ ^- ]]; then
                    plugin_name="$1"
                    shift
                fi
                ;;
            -m|--mode)
                BUILD_MODE="$2"
                shift 2
                ;;
            -j|--jobs)
                PARALLEL_JOBS="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -n|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -f|--force)
                FORCE_BUILD=true
                shift
                ;;
            -c|--clean-first)
                CLEAN_FIRST=true
                shift
                ;;
            --sequential)
                PARALLEL_JOBS=1
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Default command
    if [[ -z "$command" ]]; then
        command="build"
    fi
    
    # Validate build mode
    case "$BUILD_MODE" in
        development|production) ;;
        *)
            log_error "Invalid build mode: $BUILD_MODE"
            log_error "Valid modes: development, production"
            exit 1
            ;;
    esac
    
    # Export parsed values
    export COMMAND="$command"
    export PLUGIN_NAME="$plugin_name"
}

# Main function
main() {
    # Initialize
    init_logging
    
    # Parse arguments
    parse_args "$@"
    
    # Check prerequisites
    check_prerequisites
    
    # Execute command
    case "$COMMAND" in
        build)
            if $CLEAN_FIRST; then
                clean_plugins
            fi
            
            if [[ -n "$PLUGIN_NAME" ]]; then
                # Build specific plugin
                log_info "Building single plugin: $PLUGIN_NAME"
                if build_plugin "$PLUGIN_NAME"; then
                    log_success "Plugin build completed successfully"
                else
                    log_error "Plugin build failed"
                    exit 1
                fi
            else
                # Build all plugins
                local plugins
                mapfile -t plugins < <(discover_plugins)
                
                if [[ ${#plugins[@]} -eq 0 ]]; then
                    log_warning "No plugins to build"
                    exit 0
                fi
                
                local build_start=$(date +%s)
                
                if [[ $PARALLEL_JOBS -gt 1 ]]; then
                    build_plugins_parallel "${plugins[@]}"
                else
                    build_plugins_sequential "${plugins[@]}"
                fi
                
                local build_end=$(date +%s)
                local total_time=$((build_end - build_start))
                
                log_success "All plugins build completed in ${total_time}s"
            fi
            ;;
        clean)
            clean_plugins
            ;;
        list)
            list_plugins
            ;;
        watch)
            watch_plugins
            ;;
        help)
            show_help
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi