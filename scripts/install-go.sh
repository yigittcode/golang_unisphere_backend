#!/bin/bash
# Script to install or upgrade Go to the required version (1.23.x)

# Get the required Go version from go.mod
REQUIRED_GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
REQUIRED_GO_MAJOR=$(echo $REQUIRED_GO_VERSION | cut -d. -f1)
REQUIRED_GO_MINOR=$(echo $REQUIRED_GO_VERSION | cut -d. -f2)

# Get current Go version
CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
CURRENT_GO_MAJOR=$(echo $CURRENT_GO_VERSION | cut -d. -f1)
CURRENT_GO_MINOR=$(echo $CURRENT_GO_VERSION | cut -d. -f2)

# Compare versions
if [[ "$CURRENT_GO_MAJOR" -lt "$REQUIRED_GO_MAJOR" ]] || 
   [[ "$CURRENT_GO_MAJOR" -eq "$REQUIRED_GO_MAJOR" && "$CURRENT_GO_MINOR" -lt "$REQUIRED_GO_MINOR" ]]; then
    echo "Go version $CURRENT_GO_VERSION is installed, but $REQUIRED_GO_VERSION is required."
    read -p "Do you want to install Go $REQUIRED_GO_VERSION? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        
        # Map architecture to Go's naming
        case "$ARCH" in
            x86_64)
                ARCH="amd64"
                ;;
            aarch64|arm64)
                ARCH="arm64"
                ;;
            # Add more mappings if needed
        esac
        
        # Download the latest Go 1.23.x release
        GO_VERSION="1.23.8"  # Change this to the latest available 1.23.x version
        DOWNLOAD_URL="https://golang.org/dl/go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
        
        echo "Downloading Go $GO_VERSION from $DOWNLOAD_URL..."
        curl -L $DOWNLOAD_URL -o go.tar.gz
        
        echo "Removing previous Go installation in /usr/local/go (sudo required)..."
        sudo rm -rf /usr/local/go
        
        echo "Installing Go $GO_VERSION to /usr/local/go (sudo required)..."
        sudo tar -C /usr/local -xzf go.tar.gz
        
        echo "Cleaning up..."
        rm go.tar.gz
        
        # Ensure Go binaries are in the PATH
        if [[ ":$PATH:" != *":/usr/local/go/bin:"* ]]; then
            echo "Adding /usr/local/go/bin to PATH in your ~/.profile"
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
            echo "Please run 'source ~/.profile' to update your PATH, or log out and back in."
        fi
        
        echo "Go $GO_VERSION has been installed. Please open a new terminal or source your profile."
        echo "Run 'go version' to confirm the installation."
    else
        echo "Installation cancelled. Project requires Go $REQUIRED_GO_VERSION."
        echo "You can:"
        echo "1. Use Docker to build and run the application with 'docker-compose up'"
        echo "2. Manually install Go $REQUIRED_GO_VERSION from https://golang.org/dl/"
    fi
else
    echo "Go version $CURRENT_GO_VERSION is installed, which meets the requirement of $REQUIRED_GO_VERSION."
fi