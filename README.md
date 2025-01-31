# 🌊 Gitea Config Wave

[![Build Status](https://github.com/DUALSTACKS/gitea-config-wave/workflows/build/badge.svg)](https://github.com/DUALSTACKS/gitea-config-wave/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/DUALSTACKS/gitea-config-wave)](https://goreportcard.com/report/github.com/DUALSTACKS/gitea-config-wave)
[![License](https://img.shields.io/github/license/DUALSTACKS/gitea-config-wave)](LICENSE)
[![Release](https://img.shields.io/github/release/DUALSTACKS/gitea-config-wave.svg)](https://github.com/DUALSTACKS/gitea-config-wave/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/DUALSTACKS/gitea-config-wave)](go.mod)

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
Currently, the only way to use Gitea Config Wave is by building it from source. We're actively working on providing:

- Pre-built binaries for different platforms
- Package manager distributions (Homebrew, apt, etc.)
- Docker images
- Native packages for various Linux distributions

## Quick Start 🚀

1. First, clone and build the project:
```bash
git clone github.com/DUALSTACKS/gitea-config-wave
cd gitea-config-wave
make build
```

2. Set up your configuration in `gitea-config-wave.yaml`:
```yaml
gitea_url: ${GITEA_URL}
gitea_token: ${GITEA_TOKEN}

config:
  output_dir: .gitea/defaults

targets:
  repos: ["org/repo1", "org/repo2"]
```

3. Pull settings from your template repo:
```bash
./gitea-config-wave pull org/template-repo
```

4. Push settings to target repos:
```bash
./gitea-config-wave push --dry-run  # Preview changes
./gitea-config-wave push            # Apply changes
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
