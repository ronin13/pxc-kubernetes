Kubernetes with Percona XtraDB Cluster (PXC)
================

## Introduction

This repository contains files to spawn a cluster onto Google Cloud Engine (GCE).

The script spawn-node is idempotent in nature, you can run it as many times as you need nodes in cluster without worrying about anything else.

Multiple invocations of the script will either bootstrap a new cluster or add new nodes to cluster.

Also, a Kubernetes Service is spawned.

The script assumes that you have GCE setup correctly. Refer to [https://cloud.google.com/container-engine/docs/before-you-begin](https://cloud.google.com/container-engine/docs/before-you-begin) for more.

[https://cloud.google.com/container-engine/docs/pods/operations#pod_configuration_file](https://cloud.google.com/container-engine/docs/pods/operations#pod_configuration_file) for more details on configuration.

Also, Kubernetes documentation [here](https://github.com/GoogleCloudPlatform/kubernetes/tree/master/docs) is a valuable resource.


## Model

First, a cluster Service is spawned. After this, newer PXC node pods are added which are based on this cluster Service. Here ```Pods``` and ```Service``` refer to terms in Kubernetes parlance.


## Docker Image

The Docker image used is [https://registry.hub.docker.com/u/ronin/pxc/](https://registry.hub.docker.com/u/ronin/pxc/) with centos7 tag.
