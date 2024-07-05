# GitRelease Documentation

## Overview

GitRelease is a command-line tool designed to fetch the latest release version of a GitHub repository with ease. It simplifies the process of staying updated with the latest versions of your favorite GitHub projects.

## Installation
**Requirements**

Linux operating system (AMD64 or ARM64 architecture)
curl must be installed

### Linux Systems

To install GitRelease on Linux systems, follow these steps:

1. Download the `install.sh` script from the GitRelease repository.
2. Open a terminal and navigate to the directory where you downloaded `install.sh`.
3. Make the script executable by running `chmod +x install.sh`.
4. Execute the script as root or with sudo: `sudo ./install.sh`.

The script will automatically check for dependencies, download the latest version of GitRelease, and install it to `/usr/local/bin`.

## Usage

To use GitRelease, simply run the following command in your terminal:

```shell
gitrelease <repo-organization/repo> 
```

## Building from Source

If you prefer to build GitRelease from source, ensure you have Go installed and follow these steps:

- Clone the GitRelease repository.

- Navigate to the cloned directory.

- Run go build to compile the source code.

- The gitrelease binary will be created in the current directory.


## Support

For issues, suggestions, or contributions, please open an issue or start a discussion in the GitRelease GitHub repository.