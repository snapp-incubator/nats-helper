# nats-helper Helm Chart

A Helm chart for deploying the nats-helper service to Kubernetes.

## Introduction

This chart bootstraps a nats-helper deployment on a Kubernetes cluster using the Helm package manager.

## Prerequisites
- Kubernetes 1.16+
- Helm 3+

## Installing the Chart

Add the repository and install:

```sh
helm install my-nats-helper ./charts/nats-helper
```

## Uninstalling the Chart

```sh
helm uninstall my-nats-helper
```

## Configuration

The following table lists the configurable parameters of the nats-helper chart and their default values.

| Parameter                | Description                                 | Default             |
|--------------------------|---------------------------------------------|---------------------|
| `replicaCount`           | Number of replicas                          | `1`                 |
| `image.repository`       | Image repository                            | `snappincubator/nats-helper` |
| `image.tag`              | Image tag                                   | `latest`            |
| `image.pullPolicy`       | Image pull policy                           | `IfNotPresent`      |
| `service.type`           | Kubernetes service type                      | `ClusterIP`         |
| `service.port`           | Service port                                | `8080`              |
| `env`                    | Environment variables for the container      | `[]`                |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

## Example

```sh
helm install my-nats-helper ./charts/nats-helper \
  --set env[0].name=NATS_URL --set env[0].value=nats://nats:4222
``` 