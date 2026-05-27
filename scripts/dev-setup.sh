#!/bin/bash

# yoo Development Setup Script
# This script helps you set up the development environment for yoo

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main setup function
main() {
    print_header "yoo Development Environment Setup"

    # Check prerequisites
    print_info "Checking prerequisites..."

    MISSING_DEPS=0

    # Check Go
    if command_exists go; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go installed: $GO_VERSION"

        # Check Go version
        GO_VERSION_NUMBER=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f1,2)
        REQUIRED_VERSION="1.21"
        if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION_NUMBER" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
            print_success "Go version is sufficient (>= 1.21)"
        else
            print_warning "Go version might be too old. Recommended: 1.21+"
        fi
    else
        print_error "Go is not installed. Please install Go 1.21 or higher."
        print_info "Visit: https://golang.org/dl/"
        MISSING_DEPS=1
    fi

    # Check Git
    if command_exists git; then
        GIT_VERSION=$(git --version | awk '{print $3}')
        print_success "Git installed: $GIT_VERSION"
    else
        print_error "Git is not installed. Please install Git."
        MISSING_DEPS=1
    fi

    # Check Make (optional)
    if command_exists make; then
        print_success "Make installed"
    else
        print_warning "Make is not installed. You can still build manually with 'go build'."
    fi

    if [ $MISSING_DEPS -eq 1 ]; then
        print_error "Please install missing dependencies and run this script again."
        exit 1
    fi

    echo ""
    print_info "All required dependencies are installed!"

    # Download Go dependencies
    print_header "Downloading Dependencies"
    print_info "Running 'go mod download'..."
    if go mod download; then
        print_success "Dependencies downloaded successfully"
    else
        print_error "Failed to download dependencies"
        exit 1
    fi

    print_info "Running 'go mod tidy'..."
    if go mod tidy; then
        print_success "Dependencies tidied successfully"
    else
        print_error "Failed to tidy dependencies"
        exit 1
    fi

    # Build the project
    print_header "Building Project"
    print_info "Building yoo binary..."

    if command_exists make; then
        if make build; then
            print_success "Build successful! Binary created at: bin/yoo"
        else
            print_error "Build failed"
            exit 1
        fi
    else
        if go build -o bin/yoo main.go; then
            print_success "Build successful! Binary created at: bin/yoo"
        else
            print_error "Build failed"
            exit 1
        fi
    fi

    # Create necessary directories
    print_header "Setting Up Directories"

    # Database directory
    DB_DIR="$HOME/.local/share/yoo"
    if [ ! -d "$DB_DIR" ]; then
        mkdir -p "$DB_DIR"
        print_success "Created database directory: $DB_DIR"
    else
        print_info "Database directory already exists: $DB_DIR"
    fi

    # Config directory
    CONFIG_DIR="$HOME/.config/yoo"
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        print_success "Created config directory: $CONFIG_DIR"
    else
        print_info "Config directory already exists: $CONFIG_DIR"
    fi

    # Copy example config if it doesn't exist
    CONFIG_FILE="$CONFIG_DIR/config.yaml"
    if [ ! -f "$CONFIG_FILE" ] && [ -f "config.example.yaml" ]; then
        cp config.example.yaml "$CONFIG_FILE"
        print_success "Created config file: $CONFIG_FILE"
        print_info "You can edit this file to customize your settings"
    elif [ -f "$CONFIG_FILE" ]; then
        print_info "Config file already exists: $CONFIG_FILE"
    fi

    # Run tests
    print_header "Running Tests"
    print_info "Running test suite..."

    if go test ./... -v; then
        print_success "All tests passed!"
    else
        print_warning "Some tests failed. This might be expected if tests aren't fully implemented yet."
    fi

    # Summary
    print_header "Setup Complete!"

    echo -e "${GREEN}Your development environment is ready!${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Run the application:"
    echo "     ${BLUE}./bin/yoo${NC} or ${BLUE}make run${NC}"
    echo ""
    echo "  2. Add a test note:"
    echo "     ${BLUE}./bin/yoo add \"My first note\"${NC}"
    echo ""
    echo "  3. View your schedule:"
    echo "     ${BLUE}./bin/yoo schedule${NC}"
    echo ""
    echo "  4. Install globally (optional):"
    echo "     ${BLUE}make install${NC}"
    echo "     Then use: ${BLUE}yoo${NC}"
    echo ""
    echo "Useful commands:"
    echo "  ${BLUE}make help${NC}        - Show all available make targets"
    echo "  ${BLUE}make test${NC}        - Run tests"
    echo "  ${BLUE}make fmt${NC}         - Format code"
    echo "  ${BLUE}make build-all${NC}   - Build for all platforms"
    echo ""
    print_info "For more information, see README.md and docs/architecture.md"
    echo ""
}

# Run main function
main
