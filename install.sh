#!/usr/bin/env bash
set -e

# Capivaras
echo "Installing/Updating GitRelease..."

# Check if the user has curl
if ! [ -x "$(command -v curl)" ]; then
  echo "Error: curl is not installed. Please install curl before running this script."
  exit 1
fi

# Determine OS type and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ $OS == "darwin" ]]; then
  if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
  elif [[ $ARCH == "arm64" ]]; then
    ARCH="arm64"
  else
    echo "Error: Unsupported macOS architecture. gitrelease only supports AMD64 and ARM64 for macOS."
    exit 1
  fi
elif [[ $OS == "linux" ]]; then
  if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
  elif [[ $ARCH == "aarch64" ]]; then
    ARCH="arm64"
  else
    echo "Error: Unsupported Linux architecture. gitrelease only supports AMD64 and ARM64 for Linux."
    exit 1
  fi
else
  echo "Error: Unsupported operating system. gitrelease only supports macOS and Linux."
  exit 1
fi

# Fetch latest tag from GitHub API
LATEST_TAG=v2.2.1

echo "Latest GitRelease version is $LATEST_TAG. Downloading..."

# Define the URL and the target path
URL="https://github.com/Phillarmonic/gitrelease/releases/download/$LATEST_TAG/gitrelease-$OS-$ARCH"

TEMP_PATH="/tmp/gitrelease-$OS-$ARCH"
TARGET_PATH="/usr/local/bin/gitrelease"

echo "Downloading $URL..."

# Download the file to a temporary location
if ! curl -L -o "$TEMP_PATH" "$URL"; then
  echo "Error: Failed to download gitrelease. Please check your internet connection and try again."
  exit 1
fi

# Move the file to the target path and make it executable
echo "Installing gitrelease..."
if ! sudo mv "$TEMP_PATH" "$TARGET_PATH"; then
  echo "Error: Failed to move gitrelease to $TARGET_PATH. Do you have the necessary permissions?"
  exit 1
fi

if ! sudo chmod +x "$TARGET_PATH"; then
  echo "Error: Failed to make gitrelease executable. Do you have the necessary permissions?"
  exit 1
fi

# Verify the installation
if [[ -x "$TARGET_PATH" ]]; then
  echo "gitrelease has been successfully installed and made executable."
else
  echo "There was an error installing gitrelease."
fi