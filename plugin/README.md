# Plugin

The plugin package implements the [`plugin sdk`](https://github.com/kaytu-io/kaytu/tree/main/pkg/plugin/sdk) of the kaytu repository.

# Components

- service.go: implements [processor interface](https://github.com/kaytu-io/kaytu/blob/main/pkg/plugin/sdk/processor.go)

- [gcp](gcp/README.md): package with Google Cloud Platform components

- preferences: package with default preferences for items in processor

- [processor](processor/README.md): contains the processors for GCP plugin

- versions: default versions