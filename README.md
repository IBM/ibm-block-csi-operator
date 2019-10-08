# Operator for IBM block storage CSI driver
The Container Storage Interface (CSI) Driver for IBM block storage systems enables container orchestrators such as Kubernetes to manage the life cycle of persistent storage.

This is the official operator to deploy and manage IBM block storage CSI driver.

Supported container platforms:
  - OpenShift v4.1
  - Kubernetes v1.13

Supported operating systems:
  - RHEL 7.x (x86 architecture)
  - RHCOS 4.1 (x86 architecture)

## Prerequisites
Please see [`Prerequisites for Driver Installation`](https://github.com/IBM/ibm-block-csi-driver#prerequisites-for-driver-installation) for details.

## Installation

### Install the operator
1. Download the manifest from GitHub.
```bash
curl https://raw.githubusercontent.com/IBM/ibm-block-csi-operator/develop/deploy/ibm-block-csi-operator.yaml > ibm-block-csi-operator.yaml
```
2. (Optional): If required, update the image fields in the ibm-block-csi-operator.yaml.
3. Install the operator.

<!-- $ kubectl apply -f csi_driver.yaml  (download and install csi_driver.yaml only if you are using Kubernetes v.1.14+) -->
```bash
$ kubectl apply -f ibm-block-csi-operator.yaml
```

### Verify the operator is running:

```bash
$ kubectl get pod -l app.kubernetes.io/name=ibm-block-csi-operator -n kube-system
NAME                                    READY   STATUS    RESTARTS   AGE
ibm-block-csi-operator-5bb7996b86-xntss 2/2     Running   0          10m
```

### Create an IBMBlockCSI custom resource
1. Create an IBMBlockCSI yaml file (ibc.yaml). If required, update the repository and tag values.
```
apiVersion: csi.ibm.com/v1
kind: IBMBlockCSI
metadata:
  name: ibm-block-csi
  namespace: kube-system
spec:
  controller:
    repository: ibmcom/ibm-block-csi-driver-controller
    tag: "1.0.0"
  node:
    repository: ibmcom/ibm-block-csi-driver-node
    tag: "1.0.0"
```

2. Apply the ibc.yaml file.

```bash
$ kubectl apply -f ibc.yaml
```

### Verify the driver is running:

```bash
$ kubectl get -n kube-system pod --selector=app=ibm-block-csi-controller
NAME                         READY   STATUS    RESTARTS   AGE
ibm-block-csi-controller-0   5/5     Running   0          10m

$ kubectl get -n kube-system pod --selector=app=ibm-block-csi-node
NAME                       READY   STATUS    RESTARTS   AGE
ibm-block-csi-node-xnfgp   3/3     Running   0          10m
ibm-block-csi-node-zgh5h   3/3     Running   0          10m

```

> **Note**: For further usage details, refer to https://github.com/IBM/ibm-block-csi-driver

## Uninstalling

### 1. Delete the IBMBlockCSI custom resource.
```bash
$ kubectl delete -f ibc.yaml
```

### 2. Delete the operator.
<!-- $ kubectl delete CSIDriver ibm-block-csi-driver -->
```bash
$ kubectl delete -f deploy/ibm-block-csi-operator.yaml
```

## Licensing

Copyright 2019 IBM Corp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

