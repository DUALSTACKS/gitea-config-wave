# Kubernetes CronJob Example

This example demonstrates how to run gitea-config-wave as a nightly Kubernetes CronJob to automatically sync your Gitea repository configurations.

## Components

1. `configmap.yaml`: Contains the gitea-config-wave configuration
2. `cronjob.yaml`: The CronJob definition that runs gitea-config-wave nightly

## Setup Instructions

1. First, modify the ConfigMap in `configmap.yaml`:
   - Set your Gitea instance URL
   - Configure your organization name
   - Adjust any other settings as needed

2. Apply the Kubernetes resources:
   ```bash
   kubectl apply -f configmap.yaml
   kubectl apply -f cronjob.yaml
   ```
