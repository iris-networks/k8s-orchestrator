#!/bin/bash

# Exit if any command fails
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Installing kubectl for your GKE cluster...${NC}"

# Check OS and architecture
OS="darwin"
if [[ "$(uname)" != "Darwin" ]]; then
  OS="linux"
fi

ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
fi

echo -e "${YELLOW}Detected OS: $OS, Architecture: $ARCH${NC}"

# Get the server version if kubectl is already installed and we can connect to a cluster
SERVER_VERSION=""
if command -v kubectl &>/dev/null; then
  echo -e "${YELLOW}Checking server version...${NC}"
  SERVER_VERSION=$(kubectl version 2>/dev/null | grep "Server Version" | sed 's/.*: v\([0-9.]*\).*/\1/' || echo "")
  if [[ -n "$SERVER_VERSION" ]]; then
    echo -e "${GREEN}Detected server version: $SERVER_VERSION${NC}"
  else
    echo -e "${YELLOW}Could not detect server version. Will install latest stable version.${NC}"
    SERVER_VERSION="1.28.3" # Default to a stable version if we can't detect
  fi
else
  echo -e "${YELLOW}kubectl not found. Will install latest stable version.${NC}"
  SERVER_VERSION="1.28.3" # Default to a stable version if kubectl is not installed
fi

# Download the appropriate kubectl version
echo -e "${YELLOW}Downloading kubectl v$SERVER_VERSION...${NC}"
curl -LO "https://dl.k8s.io/release/v$SERVER_VERSION/bin/$OS/$ARCH/kubectl"

# Make it executable
chmod +x ./kubectl

# Decide where to install
INSTALL_DIR=""
if [[ -w "/usr/local/bin" ]]; then
  # We can write to /usr/local/bin without sudo
  INSTALL_DIR="/usr/local/bin"
  echo -e "${YELLOW}Installing to /usr/local/bin (no sudo required)${NC}"
  mv ./kubectl "$INSTALL_DIR/kubectl"
elif command -v sudo &>/dev/null; then
  # We have sudo, so we can write to /usr/local/bin
  INSTALL_DIR="/usr/local/bin"
  echo -e "${YELLOW}Installing to /usr/local/bin (sudo required)${NC}"
  sudo mv ./kubectl "$INSTALL_DIR/kubectl"
else
  # No sudo and can't write to /usr/local/bin, use ~/bin
  INSTALL_DIR="$HOME/bin"
  echo -e "${YELLOW}Installing to $HOME/bin${NC}"
  mkdir -p "$INSTALL_DIR"
  mv ./kubectl "$INSTALL_DIR/kubectl"
  
  # Add to PATH if needed
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Adding $INSTALL_DIR to your PATH${NC}"
    SHELL_PROFILE=""
    if [[ -f "$HOME/.zshrc" ]]; then
      SHELL_PROFILE="$HOME/.zshrc"
    elif [[ -f "$HOME/.bashrc" ]]; then
      SHELL_PROFILE="$HOME/.bashrc"
    elif [[ -f "$HOME/.bash_profile" ]]; then
      SHELL_PROFILE="$HOME/.bash_profile"
    fi
    
    if [[ -n "$SHELL_PROFILE" ]]; then
      echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_PROFILE"
      echo -e "${YELLOW}Added to $SHELL_PROFILE. Run 'source $SHELL_PROFILE' to update your current session.${NC}"
    else
      echo -e "${RED}Could not find shell profile to update.${NC}"
      echo -e "${RED}Please manually add $INSTALL_DIR to your PATH.${NC}"
    fi
  fi
fi

# Verify the installation
if [[ -n "$INSTALL_DIR" ]] && [[ -x "$INSTALL_DIR/kubectl" ]]; then
  echo -e "${GREEN}kubectl v$SERVER_VERSION installed successfully!${NC}"
  echo -e "${YELLOW}Verifying version:${NC}"

  # Use the installed kubectl directly if it's in the PATH
  if command -v kubectl &>/dev/null; then
    kubectl version --client
  else
    "$INSTALL_DIR/kubectl" version --client
  fi

  echo -e "${GREEN}Installation complete!${NC}"
  echo -e "${YELLOW}You can test the connection to your cluster with:${NC}"
  echo -e "${YELLOW}kubectl cluster-info${NC}"

  # Clean up the downloaded binary
  if [[ -f "./kubectl" ]]; then
    echo -e "${YELLOW}Cleaning up downloaded binary...${NC}"
    rm -f ./kubectl
    echo -e "${GREEN}Cleanup complete.${NC}"
  fi
else
  echo -e "${RED}Installation failed.${NC}"
  exit 1
fi