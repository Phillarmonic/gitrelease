# GitRelease
![GitHub all releases](https://img.shields.io/github/downloads/Phillarmonic/gitrelease/total)
## Overview

GitRelease is a versatile command-line tool designed to fetch the latest release version of repositories across multiple Git platforms with ease. It simplifies the process of staying updated with the latest versions of your favorite projects hosted on GitHub, GitLab, and Bitbucket. With support for token-based authentication, GitRelease enables you to access both public and private repositories seamlessly.

## Features

- **Multi-Provider Support**: Fetch latest releases from GitHub, GitLab, and Bitbucket.
- **Authentication**: Support for Personal Access Tokens and App Passwords via command-line flags or environment variables to access private repositories and increase API rate limits.
- **Flexible Usage**: Simple CLI with options to specify providers and authentication tokens.
- **Extensible**: Easily extendable to support additional Git platforms in the future.
- **Robust Error Handling**: Clear and informative error messages for various failure scenarios.

## Installation

### Requirements

- **Operating Systems**: 
  - Linux (AMD64 or ARM64 architecture)
  - MacOS (Intel or any of the new ARM64 chips)
- **Dependencies**:
  - `curl` must be installed

### Linux/Mac Systems

To install GitRelease on Linux systems, follow these steps:

1. Run this:
   
   ```bash
   curl -L https://bit.ly/gitrelease | bash
   ```

2. The script will automatically check for dependencies, download the latest version of GitRelease, and install it to `/usr/local/bin`.

If you're curious, the shortened url is:

https://raw.githubusercontent.com/Phillarmonic/gitrelease/refs/heads/master/install.sh

## Usage

GitRelease provides a straightforward command-line interface to fetch the latest release tags from supported Git providers. Below are the usage instructions and examples.  

### Basic Command

```bash
gitrelease -repo=<owner/repo>  
# Example gitrelease -repo=actions/checkout  
# will fetch the latest version of the checkout action repo on GitHub.  
# Currently it should return v4.2.0. This version can change in the future.  
```

### Fetching Latest Release by Version Prefix

In addition to fetching the latest release, GitRelease now supports fetching the latest tag that matches a specific version prefix. This is particularly useful for projects that follow semantic versioning, like the PHP repository.

To fetch the latest tag matching a version prefix, use the `-version` flag:

```bash
gitrelease -repo=<owner/repo> -provider=<github|gitlab|bitbucket> -version=<version_prefix>
```

For example, to get the latest PHP 8.3 release:

```bash
gitrelease --repo=php/php-src --provider=github --version=8.3
```

This will return the latest tag that starts with `php-8.3`, such as `php-8.3.14`.

The version prefix should be specified without any leading `v` character. GitRelease will automatically handle tags that start with `v` (e.g., `v8.2.26`) as well as those that don't (e.g., `8.2.26`).

### Authentication Flags

The documentation for authentication flags remains the same:

- `-github-token`: GitHub Personal Access Token.
- `-gitlab-token`: GitLab Personal Access Token.
- `-bitbucket-token`: Bitbucket App Password.

Alternatively, you can provide the authentication tokens via environment variables:

- `GITHUB_TOKEN`
- `GITLAB_TOKEN`
- `BITBUCKET_TOKEN`

**Note**: Command-line flags take precedence over environment variables if both are provided.

### Environment Variables

Alternatively, you can provide authentication tokens via environment variables:

- `GITHUB_TOKEN`: GitHub Personal Access Token.
- `GITLAB_TOKEN`: GitLab Personal Access Token.
- `BITBUCKET_TOKEN`: Bitbucket App Password.

**Note**: Command-line flags take precedence over environment variables if both are provided.

### Examples

1. Fetch Latest Release from GitHub (Default Provider):
   
   ```bash
   gitrelease -repo=docker/compose
   # Expected Output: v2.29.7 (the latest tag as of today 16/10/24)
   ```

2. Fetch Latest Release from GitLab:
   
   ```bash
   gitrelease -repo=gitlab-org/gitlab -provider=gitlab
   # Expected Output: v17.4.0-ee (the latest tag as of today 16/10/24)
   ```

3. Fetch Latest Release from a Private GitHub Repository Using a Token:
   
   ```bash
   gitrelease -repo=privateowner/private-repo -github-token=your_github_token
   ```

4. Fetch Latest Release from a Private GitLab Repository Using Environment Variable:
   
   ```bash
   export GITLAB_TOKEN=your_gitlab_token
   gitrelease -repo=private/namespace/project -provider=gitlab
   ```

5. Fetch Latest Tag from a Private Bitbucket Repository Using Flags:
   
   ```bash
   gitrelease -repo=privateowner/private-repo -provider=bitbucket -bitbucket-token=your_bitbucket_token
   ```

## Authentication Token Usage

GitRelease allows you to authenticate requests to access private repositories and increase API rate limits. You can provide tokens via command-line flags or environment variables.

### Providing Tokens via Flags

```bash
gitrelease -repo=<owner/repo> -provider=<provider> -github-token=<token> -gitlab-token=<token> -bitbucket-token=<token>
```

Example:

```bash
gitrelease -repo=microsoft/vscode -github-token=ghp_yourgithubtoken
```

### Providing Tokens via Environment Variables

Set the appropriate environment variable before running GitRelease.

```bash
export GITHUB_TOKEN=ghp_yourgithubtoken
export GITLAB_TOKEN=glpat_yourgitlabtoken
export BITBUCKET_TOKEN=your_bitbuckettoken

gitrelease -repo=gitlab-org/gitlab -provider=gitlab
```

**Note**: If both flags and environment variables are set for a token, flags take precedence.

## Building from Source

If you prefer to build GitRelease from source, ensure you have Go installed and follow these steps:

1. Clone the GitRelease Repository:
   
   ```bash
   git clone https://github.com/phillarmonic/gitrelease.git
   ```

2. Navigate to the Cloned Directory:
   
   ```bash
   cd gitrelease
   ```

3. Build the Executable:
   
   ```bash
   go build -o gitrelease
   ```
   
   The `gitrelease` binary will be created in the current directory.

4. Move the Binary to a Directory in Your PATH:
   
   ```bash
   sudo mv gitrelease /usr/local/bin/
   ```

## Support

For issues, suggestions, or contributions, please open an issue or start a discussion in the GitRelease GitHub repository.