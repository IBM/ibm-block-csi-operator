# IBM Storage Orchestration Operator for IBM Block Storage
The Container Storage Interface (CSI) Driver for IBM block storage systems enables container orchestrators such as Kubernetes to manage the life-cycle of persistent storage.

This is the official operator to deploy and manage the CSI Driver for IBM block storage systems.

Supported container platforms:
  - Openshift v4.1
  - Kubernetes v1.13

Supported IBM storage systems:
  - IBM FlashSystem 9100
  - IBM Spectrum Virtualize
  - IBM Storwize
  - IBM FlashSystem A9000\R

Supported operating systems:
  - RHEL 7.x (x86 architecture)

DISCLAIMER: The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.


## Prerequisite

### Worker nodes preparation
Perform these steps for each worker node in Kubernetes cluster:

### 1. Install Linux packages to ensure Fibre Channel and iSCSI connectivity
Skip this step, if the packages are already installed.

RHEL 7.x:
```bash
$ yum -y install sg3_utils
$ yum -y install iscsi-initiator-utils   # only if iSCSI connectivity is required
$ yum -y install xfsprogs                # Only if xfs filesystem is required.
```

#### 2. Configure Linux multipath devices on the host.
Create and set the relevant storage system parameters in the `/etc/multipath.conf` file.
You can also use the default `multipath.conf` file located in the `/usr/share/doc/device-mapper-multipath-*` directory.
Verify that the `systemctl status multipathd` output indicates that the multipath status is active and error-free.

RHEL 7.x:
```bash
$ yum install device-mapper-multipath
$ modprobe dm-multipath
$ systemctl start multipathd
$ systemctl status multipathd
$ multipath -ll
```

Important: When configuring Linux multipath devices, verify that the `find_multipaths` parameter in the `multipath.conf` file is disabled.
  - RHEL 7.x: Remove the `find_multipaths yes` string from the `multipath.conf` file.

#### 3. Configure storage system connectivity.
3.1. Define the hostname of each Kubernetes node on the relevant storage systems with the valid WWPN or IQN of the node.

3.2. For Fiber Chanel, configure the relevant zoning from the storage to the host.

3.3. For iSCSI, perform these three steps.

3.3.1. Make sure that the login used to log in to the iSCSI targets is permanent and remains available after a reboot of the worker node. To do this, verify that the node.startup in the /etc/iscsi/iscsid.conf file is set to automatic. If not, set it as required and then restart the iscsid service `$> service iscsid restart`.

3.3.2. Discover and log into at least two iSCSI targets on the relevant storage
systems.

```bash
$ iscsiadm -m discoverydb -t st -p ${storage system iSCSI port IP}:3260
--discover
$ iscsiadm -m node -p ${storage system iSCSI port IP/hostname} --login
```

3.3.3. Verify that the login was successful and display all targets that you logged in. The portal value must be the iSCSI target IP address.

```bash
$ iscsiadm -m session --rescan
Rescanning session [sid: 1, target: {storage system IQN},
portal: {storage system iSCSI port IP},{port number}
```

End of worker node setup.


### Install CSIDriver CRD - optional
Enabling CSIDriver on Kubernetes (more details -> https://kubernetes-csi.github.io/docs/csi-driver-object.html#enabling-csidriver-on-kubernetes)

In Kubernetes v1.13, because the feature was alpha, it was disabled by default. To enable the use of CSIDriver on these versions, do the following:

1. Ensure the feature gate is enabled via the following Kubernetes feature flag: --feature-gates=CSIDriverRegistry=true
   For example on kubeadm installation add the flag inside the `/etc/kubernetes/manifests/kube-apiserver.yaml`.
2. Either ensure the CSIDriver CRD is automatically installed via the Kubernetes Storage CRD addon OR manually install the CSIDriver CRD on the Kubernetes cluster with the following command:
   ```bash
   $ kubectl create -f https://raw.githubusercontent.com/kubernetes/csi-api/master/pkg/crd/manifests/csidriver.yaml
   ```

If the feature gate was not enabled then CSIDriver for the ibm-block-csi-driver will not be created automatically.


<br/>

## Installation

### Install the CSI driver operator

#### Install with helm

```bash
$ helm repo add artifactory https://stg-artifactory.haifa.ibm.com/artifactory/chart-repo
$ helm install --name ibm-block-csi-driver-operator --namespace kube-system artifactory/ibm-block-csi-driver-operator
```

#### Install with yaml

```bash

$ kubectl apply -f deploy/csi_driver.yaml  (install csi_driver.yaml only if you are using Kubernetes v.1.14+)
$ kubectl apply -f deploy/ibm-block-csi-driver-operator.yaml
```

### Verify operator is running:

```bash
$ kubectl get pod -l app.kubernetes.io/name=ibm-block-csi-driver-operator -n kube-system
NAME                                             READY   STATUS    RESTARTS   AGE
ibm-block-csi-driver-operator-5bb7996b86-xntss   2/2     Running   0          10m
```

### Create an IBMBlockCSI custom resource
Create an IBMBlockCSI yaml file (ibc.yaml) as following and update the relevant fields:
```
apiVersion: csi.ibm.com/v1
kind: IBMBlockCSI
metadata:
  name: ibm-block-csi
  namespace: kube-system
spec:
  controller:
    repository: stg-artifactory.haifa.ibm.com:5030/ibm-block-csi-controller-driver
    tag: "SVC_Storage_Action"
  node:
    repository: stg-artifactory.haifa.ibm.com:5030/ibm-block-csi-node-driver
    tag: "1.0.0_b74_origin.feature.csi-node-daemonset"
```

Apply it:

```bash
$ kubectl apply -f ibc.yaml
```

### Verify driver is running:

```bash
$ kubectl get -n kube-system pod --selector=app=ibm-block-csi-controller
NAME                         READY   STATUS    RESTARTS   AGE
ibm-block-csi-controller-0   5/5     Running   0          10m

$ kubectl get -n kube-system pod --selector=app=ibm-block-csi-node
NAME                       READY   STATUS    RESTARTS   AGE
ibm-block-csi-node-xnfgp   3/3     Running   0          10m
ibm-block-csi-node-zgh5h   3/3     Running   0          10m

```

> **Note**: For further usage, please go to https://github.com/IBM/ibm-block-csi-driver

## Uninstallation

### 1. Delete the IBMBlockCSI custom resource
```bash
$ kubectl delete -f ibc.yaml
```


### 2. Delete the operator

#### Delete with helm
```bash
$ helm delete --purge ibm-block-csi-driver-operator
```
#### Delete with yaml
```bash
$ kubectl delete CSIDriver ibm-block-csi-driver
$ kubectl delete -f deploy/ibm-block-csi-driver-operator.yaml
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

