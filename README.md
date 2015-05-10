Kubernetes with Percona XtraDB Cluster (PXC) on Google Cloud Engine (GCE)
================

## Introduction

This repository contains files to spawn a PXC cluster onto Google Cloud Engine (GCE).

## How it runs

* The script spawn-node is idempotent in nature, you can run it as many times as you need nodes in cluster without worrying about anything else. 

* Multiple invocations of the script will either bootstrap a new cluster or add new nodes to cluster through addition of new ```Pods```.

* Also, a Kubernetes cluster ```Service``` is spawned.

## Setup

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
```bash


bash -x ./spawn.sh
+ trap cleanup EXIT
+ gcloud alpha container kubectl get services -l type=cluster
+ grep -q cluster
++ gcloud alpha container kubectl get pods -l name=pxc
++ wc -l
+ count=2
+ count=0
+ nodecnt=0
+ [[ 0 -eq 0 ]]
+ bstrap='"--wsrep-new-cluster",'
+ joiner=
+ source template
+ gcloud alpha container kubectl create -f -
+ tee -a /tmp/log
++ cat
pods/node0

+ cleanup
+ local estatus=0
+ [[ 0 -ne 0 ]]
                                                                                                                                                                                                                                  (google-cloud/eternal-autumn-94011)~21:58-0
>>        gcloud alpha container kubectl get pods -l 'name=pxc'
POD       IP            CONTAINER(S)   IMAGE(S)            HOST                                      LABELS     STATUS    CREATED
node0     10.140.1.19   node0          ronin/pxc:centos7   k8s-galera-cluster-node-1/104.197.66.82   name=pxc   Running   Less than a second



./spawn.sh
+ trap cleanup EXIT
+ gcloud alpha container kubectl get services -l type=cluster
+ grep -q cluster
++ gcloud alpha container kubectl get pods -l name=pxc
++ wc -l
+ count=3
+ count=1
+ nodecnt=1
+ [[ 1 -eq 0 ]]
+ bstrap=
+ joiner=cluster
+ source template
+ gcloud alpha container kubectl create -f -
++ cat
+ tee -a /tmp/log
pods/node1

+ cleanup
+ local estatus=0
+ [[ 0 -ne 0 ]]
+ gcloud alpha container kubectl get pods -l name=pxc
POD       IP            CONTAINER(S)   IMAGE(S)            HOST                                      LABELS     STATUS    CREATED
node0     10.140.1.19   node0          ronin/pxc:centos7   k8s-galera-cluster-node-1/104.197.66.82   name=pxc   Running   Less than a second
node1     10.140.1.20   node1          ronin/pxc:centos7   k8s-galera-cluster-node-1/104.197.66.82   name=pxc   Running   Less than a second


```
