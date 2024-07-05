#!/usr/bin/env bash

set -eu

# Checks if the is running as sudo
if [[ $EUID -ne 0 ]]; then
  echo "This script must be run as root. Please run this command as sudo or root user."
  exit 1
fi

# Check if the user has curl
if ! [ -x "$(command -v curl)" ]; then
  echo "Error: curl is not installed. Please install curl before running this script."
  exit 1
fi

# Determine if OS architecture is AMD64 or ARM64
ARCH=$(uname -m)

if [[ $ARCH == "x86_64" ]]; then
  ARCH="amd64"
elif [[ $ARCH == "aarch64" ]]; then
  ARCH="arm64"
else
  echo "Error: Unsupported architecture. gitrelease only supports AMD64 and ARM64."
  exit 1
fi

# Fetch latest tag from GitHub api
LATEST_TAG=$(curl -s "https://api.github.com/repos/Phillarmonic/gitrelease/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')
echo "Latest GitRelease version is $LATEST_TAG. Downloading..."
# Define the URL and the target path
URL="https://github.com/Phillarmonic/gitrelease/releases/download/$LATEST_TAG/gitrelease-linux-$ARCH"

TEMP_PATH="/tmp/gitrelease-linux-$ARCH"
TARGET_PATH="/usr/local/bin/gitrelease"

echo "Downloading $URL..."

# Download the file to a temporary location
curl -L "$URL" -o "$TEMP_PATH"

# Move the file to the target path and make it executable
echo "Installing gitrelease..."
sudo mv "$TEMP_PATH" $TARGET_PATH
sudo chmod +x $TARGET_PATH

# Verify the installation
if [[ -x $TARGET_PATH ]]; then
  echo "gitrelease has been successfully installed and made executable."
else
  echo "There was an error installing gitrelease."
fi
