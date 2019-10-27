# Operator for IBM block storage CSI driver
The Container Storage Interface (CSI) Driver for IBM block storage systems enables container orchestrators such as Kubernetes to manage the life cycle of persistent storage.

This is the official operator to deploy and manage IBM block storage CSI driver.

Supported container platforms:
  - OpenShift v4.2+
  - Kubernetes v1.14+

Supported operating systems:
  - RHEL 7.x (x86 architecture)
  - RHCOS 4.1 (x86 architecture)

## Prerequisites
Please see [`Prerequisites for Driver Installation`](https://github.com/IBM/ibm-block-csi-driver#prerequisites-for-driver-installation) for details.

## Installation

### Install the operator

> **Note**: The operator can be installed directly from the RedHat OpenShift web console, through the OperatorHub.


1. Download the manifest from GitHub.
```bash
curl https://raw.githubusercontent.com/IBM/ibm-block-csi-operator/master/deploy/installer/generated/ibm-block-csi-operator.yaml > ibm-block-csi-operator.yaml
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


1. Download the manifest from GitHub.
```bash
curl https://raw.githubusercontent.com/IBM/ibm-block-csi-operator/master/deploy/crds/csi.ibm.com_v1_ibmblockcsi_cr.yaml > csi.ibm.com_v1_ibmblockcsi_cr.yaml
```

2. (Optional): If required, update the image fields in the csi.ibm.com_v1_ibmblockcsi_cr.yaml.

3. Install the csi.ibm.com_v1_ibmblockcsi_cr.yaml.

<!-- $ kubectl apply -f csi.ibm.com_v1_ibmblockcsi_cr.yaml -->
```bash
$ kubectl apply -f csi.ibm.com_v1_ibmblockcsi_cr.yaml
```

### Verify the driver is running:

```bash
$> kubectl get all -n kube-system  -l csi
NAME                             READY   STATUS    RESTARTS   AGE
pod/ibm-block-csi-controller-0   5/5     Running   0          9m36s
pod/ibm-block-csi-node-jvmvh     3/3     Running   0          9m36s
pod/ibm-block-csi-node-tsppw     3/3     Running   0          9m36s

NAME                                DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/ibm-block-csi-node   2         2         2       2            2           <none>          9m36s

NAME                                        READY   AGE
statefulset.apps/ibm-block-csi-controller   1/1     9m36s
```

> **Note**: For further usage details, refer to https://github.com/IBM/ibm-block-csi-driver

## Uninstalling

### 1. Delete the IBMBlockCSI custom resource.
```bash
$ kubectl delete -f csi.ibm.com_v1_ibmblockcsi_cr.yaml
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

