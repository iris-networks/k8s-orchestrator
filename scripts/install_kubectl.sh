#!/bin/bash

# Exit if any command fails
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Installing kubectl and GKE auth plugin for your macOS GKE setup...${NC}"

# Check if running on macOS
if [[ "$(uname)" != "Darwin" ]]; then
  echo -e "${RED}This script is designed for macOS only.${NC}"
  exit 1
fi

# Get macOS architecture
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
fi

echo -e "${YELLOW}Detected macOS with $ARCH architecture${NC}"

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
echo -e "${YELLOW}Downloading kubectl v$SERVER_VERSION for macOS $ARCH...${NC}"
curl -LO "https://dl.k8s.io/release/v$SERVER_VERSION/bin/darwin/$ARCH/kubectl"

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

# Verify the kubectl installation
if [[ -n "$INSTALL_DIR" ]] && [[ -x "$INSTALL_DIR/kubectl" ]]; then
  echo -e "${GREEN}kubectl v$SERVER_VERSION installed successfully!${NC}"
  echo -e "${YELLOW}Verifying kubectl version:${NC}"
  
  # Use the installed kubectl directly if it's in the PATH
  if command -v kubectl &>/dev/null; then
    kubectl version --client
  else
    "$INSTALL_DIR/kubectl" version --client
  fi
  
  # Clean up the downloaded kubectl binary
  if [[ -f "./kubectl" ]]; then
    echo -e "${YELLOW}Cleaning up downloaded kubectl binary...${NC}"
    rm -f ./kubectl
    echo -e "${GREEN}kubectl cleanup complete.${NC}"
  fi
else
  echo -e "${RED}kubectl installation failed.${NC}"
  exit 1
fi

# Install GKE auth plugin
echo -e "${YELLOW}Checking for gke-gcloud-auth-plugin...${NC}"
if command -v gke-gcloud-auth-plugin &>/dev/null; then
  echo -e "${GREEN}gke-gcloud-auth-plugin is already installed.${NC}"
  gke-gcloud-auth-plugin --version
else
  echo -e "${YELLOW}gke-gcloud-auth-plugin not found. Installing now...${NC}"

  # Check if gcloud is available
  if command -v gcloud &>/dev/null; then
    echo -e "${YELLOW}Installing gke-gcloud-auth-plugin via gcloud...${NC}"
    gcloud components install gke-gcloud-auth-plugin

    if command -v gke-gcloud-auth-plugin &>/dev/null; then
      echo -e "${GREEN}gke-gcloud-auth-plugin installed successfully!${NC}"
      gke-gcloud-auth-plugin --version
    else
      echo -e "${RED}Failed to install gke-gcloud-auth-plugin via gcloud.${NC}"
      echo -e "${YELLOW}Please install it manually using Homebrew:${NC}"
      echo -e "${YELLOW}  brew install --cask google-cloud-sdk${NC}"
      echo -e "${YELLOW}  gcloud components install gke-gcloud-auth-plugin${NC}"
    fi
  else
    echo -e "${RED}gcloud command not found. Cannot install gke-gcloud-auth-plugin automatically.${NC}"
    echo -e "${YELLOW}Please install gcloud SDK and the GKE auth plugin manually:${NC}"
    echo -e "${YELLOW}  brew install --cask google-cloud-sdk${NC}"
    echo -e "${YELLOW}  gcloud components install gke-gcloud-auth-plugin${NC}"
  fi
fi

echo -e "${GREEN}Installation complete!${NC}"
echo -e "${YELLOW}You can test the connection to your cluster with:${NC}"
echo -e "${YELLOW}kubectl cluster-info${NC}"