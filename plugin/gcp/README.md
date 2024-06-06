# GCP package for wrapping GCP api

## About each component

- GCP
    - Base configuration for all Google Cloud Clients 
    - Embedded by other clients e.g. Compute Instance types

- Compute
    - Controls Compute client for GCP
    - Gets list of Instances

- Metrics
    - Controls Metrics client for GCP
    - 



TODO:

 - Metrics Client
    - Close metrics client

 - Metrics to collect:
    - "CPUUtilization",

    - "NetworkIn",
    - "NetworkOut",

    - "mem_used_percent",

    - "VolumeReadBytes",
    - "VolumeWriteBytes",

    - "VolumeReadOps",
    - "VolumeWriteOps",