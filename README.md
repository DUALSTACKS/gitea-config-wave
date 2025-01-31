# 🌊 Gitea Config Wave

<div align="center">
  
[![Build Status](https://github.com/DUALSTACKS/gitea-config-wave/workflows/build/badge.svg)](https://github.com/DUALSTACKS/gitea-config-wave/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/DUALSTACKS/gitea-config-wave)](https://goreportcard.com/report/github.com/DUALSTACKS/gitea-config-wave)
[![License](https://img.shields.io/github/license/DUALSTACKS/gitea-config-wave)](LICENSE)
[![Release](https://img.shields.io/github/release/DUALSTACKS/gitea-config-wave.svg)](https://github.com/DUALSTACKS/gitea-config-wave/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/DUALSTACKS/gitea-config-wave)](go.mod)

</div>

<div align="center">
  <img src="logo.png" alt="Gitea Config Wave Logo" width="300"/>
</div>

<br/>

> Effortlessly enforce branch protections, standardize external trackers, and maintain consistent repository defaults across your entire self-hosted Gitea instance with just one command - say goodbye to manual configuration! 🚀

## Project Status 🚧

This project is in early development - use with caution in production environments.

## What is this? 🤔

Gitea Config Wave is a CLI tool that helps you manage repository settings across multiple repositories in your Gitea instance. Think of it as a "settings propagator" that lets you define settings once and apply them everywhere!

## The Problem 🗿

If you're managing many repos in a self-hosted Gitea instance (not Enterprise), you might have faced these challenges:
- No way to set organization-wide branch protection rules or repo defaults like external trackers
- Manual configuration needed for each repository's settings
- Time-consuming process to maintain consistent settings across repos
- Need to push issue and PR templates to each repo

## The Solution 🎯

Gitea Config Wave provides a simple CLI to pull settings from a "canonical" repository and push them to multiple target repositories. It's like copy-paste, but for repo settings!

## Features ✨

- 📥 **Pull Settings**: Extract settings from any repo to use as a template
- 🔎 **Manage config as YAML**: Store and version control your repository settings as YAML files
- 📤 **Push Settings**: Apply settings to multiple repos at once
- 🛡️ **Branch Protection**: Sync branch protection rules across repos
- 🎯 **Repository Settings**: Manage core repo settings and topics
- 🔍 **Dry Run Mode**: Preview changes before applying them
- 🤖 **Automation Ready**: Perfect for CI/CD pipelines

## Installation 🔧

Choose one of the following installation methods:

### Using Homebrew (macOS/Linux)

```bash
# Install
brew install dualstacks/tap/gitea-config-wave

# Upgrade
brew upgrade gitea-config-wave
```

### Using Pre-built Binaries

1. Download the latest binary for your platform from the [releases page](https://github.com/DUALSTACKS/gitea-config-wave/releases)
2. Extract the archive (if applicable)
3. Move the binary to your PATH:
```bash
# Example for macOS/Linux
chmod +x gitea-config-wave
sudo mv gitea-config-wave /usr/local/bin/
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/DUALSTACKS/gitea-config-wave.git
cd gitea-config-wave

# Build the binary
make build

# Optional: Install to your PATH
sudo make install
```

## Quick Start 🚀

### 1. Create a Gitea Access Token

1. Log in to your Gitea instance
2. Go to Settings → Applications → Generate New Token
3. Give your token a name (e.g., "Config Wave")
4. Select the following permissions:
   - `read:organization`: Read organization information
   - `write:repository`: Full control of repositories (includes settings, webhooks, and branch protections)
5. Click "Generate Token" and save it securely

### 2. Configure the Tool

Create a `gitea-config-wave.yaml` in your working directory:

```yaml
gitea_url: "https://your-gitea-instance.com"  # Your Gitea instance URL
gitea_token: "${GITEA_TOKEN}"                 # Use environment variable for token

config:
  output_dir: .gitea/defaults                 # Where to store pulled settings

targets:
  repos:
    - "org/repo1"                            # List of target repositories
    - "org/repo2"
    - "org/repo3"
```

### 3. Pull Settings from a Template

```bash
# Export your Gitea token
export GITEA_TOKEN="your-token-here"

# Pull settings from a template repository
gitea-config-wave pull org/template-repo

# This will create YAML files in .gitea/defaults/ with the current settings
```

### 4. Review and Customize Settings

The pulled settings are stored in YAML files:
- `.gitea/defaults/branch_protections.yaml`: Branch protection rules
- `.gitea/defaults/repo_settings.yaml`: Repository settings
- `.gitea/defaults/topics.yaml`: Repository topics
- `.gitea/defaults/webhooks.yaml`: Webhook configurations

### 5. Push Settings to Target Repositories

```bash
# Preview changes (dry run)
gitea-config-wave push --dry-run

# Apply changes to all target repositories
gitea-config-wave push
```

## Configuration Examples 📝

### Branch Protection Rules

```yaml
# .gitea/defaults/branch_protections.yaml
branch_protections:
  - branch_name: "main"
    enable_push: false
    enable_push_whitelist: true
    push_whitelist_usernames: ["maintainer1", "maintainer2"]
    enable_status_check: true
    status_check_contexts: ["ci/jenkins"]
    require_signed_commits: true
```

### Repository Settings

```yaml
# .gitea/defaults/repo_settings.yaml
repository_settings:
  enable_issues: true
  enable_projects: true
  enable_pull_requests: true
  ignore_whitespace_conflicts: true
  enable_merge_commits: false
  enable_rebase: true
  enable_squash: true
  default_merge_style: "rebase"
```

## Configuration 🛠️

The tool expects a `gitea-config-wave.yaml` file in the current directory. Refer to the [example configuration](./gitea-config-wave.yaml) for more details.

## Use Cases 💡

- Setting up consistent branch protection rules across repos
- Configuring external issue trackers (e.g., Jira) for multiple repos
- Maintaining uniform PR and issue templates
- Enforcing organization-wide repository settings

## Project Status 🚧

This project is in early development - use with caution in production environments.

## Backlog 📝

- [x] Webhook management
- [x] Basic settings sync
- [x] Branch protection rules
- [ ] Issue and PR template sync
- [ ] Tag protection rules
- [ ] Distribute as Gitea Action


## Contributing 🤝

TODO:

## License 📝

MIT License - see [LICENSE](LICENSE) for details.

---

Gitea Config Wave - An automated configuration management tool for Gitea instances that enables bulk repository settings management, branch protection rules enforcement, and standardized repository configurations through a simple CLI interface.


<details>
<summary></summary>
<small>
gitea organization-wide branch protection rules; gitea organization-wide repository settings; gitea bulk repository settings; gitea repository settings sync; bulk edit gitea topics; bulk edit gitea webhooks; bulk edit gitea branch protections; bulk edit gitea repo settings;
</small>
</details>
