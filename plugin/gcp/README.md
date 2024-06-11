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
    - "CPUUtilization" ~> `compute.googleapis.com/instance/cpu/utilization`

    - "NetworkIn" ~> `compute.googleapis.com/instance/network/received_bytes_count`
    - "NetworkOut" ~> `compute.googleapis.com/instance/network/sent_bytes_count`

    - "mem_used_percent" ~> `compute.googleapis.com/instance/memory/balloon/ram_size`/`compute.googleapis.com/instance/memory/balloon/ram_used`

    - "VolumeReadBytes" ~> `compute.googleapis.com/instance/disk/read_bytes_count`
    - "VolumeWriteBytes" ~> `compute.googleapis.com/instance/disk/write_bytes_count`

    - "VolumeReadOps" ~> `compute.googleapis.com/instance/disk/read_ops_count`
    - "VolumeWriteOps" ~> `compute.googleapis.com/instance/disk/write_ops_count`