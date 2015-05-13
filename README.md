Kubernetes with Percona XtraDB Cluster (PXC) on Google Cloud Engine (GCE)
================

## Introduction

This repository contains files to spawn a PXC cluster onto Google Cloud Engine (GCE).

## How it runs

* The script kubecluster is idempotent in nature, you can run it as many times as you need nodes in cluster without worrying about anything else. 

* Multiple invocations of the script will either bootstrap a new cluster or add new nodes to cluster through addition of new ```Pods```.

* Also, a Kubernetes cluster ```Service``` is spawned.

## Setup

* Make sure to install the latest Google Cloud SDK. The latest at this time is [0.9.60](https://gist.github.com/61b2f6cb8aefdf31dbee).

* The script assumes that you have GCE setup correctly. Refer to [https://cloud.google.com/container-engine/docs/before-you-begin](https://cloud.google.com/container-engine/docs/before-you-begin) for more.

* [https://cloud.google.com/container-engine/docs/pods/operations#pod_configuration_file](https://cloud.google.com/container-engine/docs/pods/operations#pod_configuration_file) for more details on configuration.

* Also, Kubernetes documentation [here](https://github.com/GoogleCloudPlatform/kubernetes/tree/master/docs) is a valuable resource. Make sure to understand concept of ```Pods```, ```Services``` and others as described [here](https://github.com/GoogleCloudPlatform/kubernetes/blob/master/docs/user-guide.md).

* Make sure Google cloud SDK is installed for gcloud and other CLI utils. Also, make sure to spawn a resonably large instance of GCE since memory is often consumed by other instances/services (DNS etc), so memory available for mysqld will be less. I have tested with n1-standard-1  instance.


## Model

* First, a cluster Service is spawned. After this, newer PXC node pods are added which are based on this cluster Service. Here ```Pods``` and ```Service``` refer to terms in Kubernetes parlance.

* The json configuration for a PXC pod is dynamically generated and fed to gcloud. This is required since few parameters such as wsrep_cluster_address and wsrep_node_name need to be dynamically generated. This may be removed in future if Kubernetes service can block till at least one node is up.

* All individual PXC nodes live in separate pods.

* They communicate through the cluster ```Service```. In other words, the gcomm url is ```gcomm://cluster``` which points to the endpoint of cluster ```Service``` and not bootstrapped node's IP Address. This address agnostic approach is useful in many ways and makes it easy to scale.

## Networking

* There is no Docker-like linking done here, though there seems to be a syntactical support for it.

* Instead, ```Service``` endpoints are used for communication among ```Pods```. Note that each node is in a ```Pod``` of its own.

* In a cluster, each node is both a client and a server to this ```Service``` and its endpoint.

* Using a Service also provides for load-balancing among the nodes of cluster.

## Docker Image

* The Docker image used is [https://registry.hub.docker.com/u/ronin/pxc/](https://registry.hub.docker.com/u/ronin/pxc/) with centos7 tag. This images installs a centos7 release build (may not be latest, yet, since there are no automated builds) of PXC.

## Future Work

* Addition of others into PXC pod - a haproxy or a xinetd check perhaps.

* Allow for volumes and environment for further user customization.

* Replication controllers. 

* Tests along lines of [capem](https://github.com/ronin13/capem).

## Example 

### Creating cluster
```bash

            go run kubecluster.go -create -name=testx --project=eternal-autumn-94011
            2015/05/13 23:47:22 Lets begin
            2015/05/13 23:47:22 Creating cluster testx in zone us-central1-a
            2015/05/13 23:47:22 Running gcloud alpha container clusters create testx --zone us-central1-a
            Creating cluster testx...done.
            Created [https://www.googleapis.com/container/v1beta1/projects/eternal-autumn-94011/zones/us-central1-a/clusters/testx].
            Warning: Permanently added '146.148.81.41' (ECDSA) to the list of known hosts.
            kubeconfig entry generated for testx. To switch context to the cluster, run

            $ kubectl config use-context gke_eternal-autumn-94011_us-central1-a_testx

            2015/05/13 23:51:20 Using gcloud compute copy-files to fetch ssl certs from cluster master...
            NAME   ZONE           CLUSTER_API_VERSION  MASTER_IP      MACHINE_TYPE                           NODES  STATUS
            testx  us-central1-a  0.17.0               146.148.81.41  n1-standard-1, container-vm-v20150505  3      running

            2015/05/13 23:51:20 Running gcloud config set container/cluster testx
            2015/05/13 23:51:21 Running kubectl config use-context gke_eternal-autumn-94011_us-central1-a_testx
            2015/05/13 23:51:21 Running gcloud alpha container clusters describe testx
```

### Starting a node
```bash
            go run kubecluster.go    --project=eternal-autumn-94011 -start
            2015/05/14 00:05:41 Lets begin
            2015/05/14 00:05:41 Running kubectl get services  -l 'type=cluster'
            2015/05/14 00:05:43 Running kubectl get pods --no-headers=true  -l 'name=pxc' | wc -l
            2015/05/14 00:05:44 0 nodes are up
            2015/05/14 00:05:44 Starting node0 with following configuration

            {
            "id": "node0",
            "kind": "Pod",
            "apiVersion": "v1beta1",
            "desiredState": {
            "manifest": {
                "version": "v1beta1",
                "id": "node0",
                "containers": [{
                "name": "node0",
                "image": "ronin/pxc:centos7-release",
                "ports": [{ "containerPort": 3306 }, {"containerPort": 4567 }, {"containerPort": 4568 } ],
                "command": ["/usr/sbin/mysqld",  "--basedir=/usr",  "--wsrep-node-name=node0",   "--user=mysql", "--wsrep-new-cluster",  "--skip-grant-tables", "--wsrep_cluster_address=gcomm://", "--wsrep-sst-method=rsync"]
                }]
            }
            },
            "labels": {
                "name": "pxc"
            }
            }
            pods/node0
            2015/05/14 00:05:50 Successfully started node0
            2015/05/14 00:05:50 Running kubectl get pods -l 'name=pxc'
            2015/05/14 00:05:51 POD       IP        CONTAINER(S)   IMAGE(S)                    HOST                LABELS     STATUS    CREATED     MESSAGE
            node0                                                          k8s-testx-node-1/   name=pxc   Pending   4 seconds
                                node0          ronin/pxc:centos7-release
```
