# Multi-Target Config Example

This example demonstrates managing different configurations for production services versus prototyping repositories:

1. **Production Services** (using `production-config.yaml`):
   - Strict branch protections with required reviews and multiple checks
   - Clean git history with squash-only merging
   - Protected critical files
   - JIRA integration
   - Full CI/CD pipeline requirements
   - Located in `configs/production/`

2. **Prototyping Projects** (using `prototyping-config.yaml`):
   - Basic branch protections
   - Flexible merge options for experimentation
   - Minimal CI requirements
   - Linear issue tracker integration
   - Located in `configs/prototyping/`

## Usage

Each configuration needs to be applied separately using the `--config` flag:

```bash
# Push configs to production services
gitea-config-wave push --config production-config.yaml

# Push configs to prototyping repositories
gitea-config-wave push --config prototyping-config.yaml
```
