# Partial Config Example

You can choose to update only specific repository settings while preserving all other settings. In `partial-config-example.yaml`, only `default_branch` and `has_wiki` are specified - all other repository settings will remain unchanged when pushing this config.

See `defaults/repo_settings.yaml` for all available settings.

To run this example, use the following command:

```bash
gitea-config-wave push --config partial-config-example.yaml
```
