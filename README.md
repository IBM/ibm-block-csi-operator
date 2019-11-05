# Operator for IBM block storage CSI driver
The Container Storage Interface (CSI) Driver for IBM block storage systems enables container orchestrators such as Kubernetes to manage the life cycle of persistent storage.

This is the official operator to deploy and manage IBM block storage CSI driver.

Supported container platforms:
  - OpenShift v4.2
  - Kubernetes v1.14

Supported IBM storage systems:
  - IBM FlashSystem 9100
  - IBM Spectrum Virtualize
  - IBM Storwize
  - IBM FlashSystem A9000/R

Supported operating systems:
  - RHEL 7.x (x86 architecture)

Full documentation can be found on the [IBM knowledge center](www.ibm.com/support/knowledgecenter/SSRQ8T).

<br/>
<br/>
<br/>

## Prerequisites
> **Note**: The operator can be installed directly from the RedHat OpenShift web console, through the OperatorHub. The prerequisites below also mentioned and can be viewed via OpenShift. 

### Preparing worker nodes
Perform these steps for each worker node in Kubernetes cluster:

#### 1. Install Linux packages to ensure Fibre Channel and iSCSI connectivity
Skip this step if the packages are already installed.

RHEL 7.x:
```bash
yum -y install iscsi-initiator-utils   # Only if iSCSI connectivity is required
yum -y install xfsprogs                # Only if XFS file system is required
```

#### 2. Configure Linux multipath devices on the host 
Create and set the relevant storage system parameters in the `/etc/multipath.conf` file. 
You can also use the default `multipath.conf` file, located in the `/usr/share/doc/device-mapper-multipath-*` directory.
Verify that the `systemctl status multipathd` output indicates that the multipath status is active and error-free.

RHEL 7.x:
```bash
yum install device-mapper-multipath
modprobe dm-multipath
systemctl enable multipathd
systemctl start multipathd
systemctl status multipathd
multipath -ll
```

**Important:** When configuring Linux multipath devices, verify that the `find_multipaths` parameter in the `multipath.conf` file is disabled. In RHEL 7.x, remove the`find_multipaths yes` string from the `multipath.conf` file.

#### 3. Configure storage system connectivity
3.1. Define the hostname of each Kubernetes node on the relevant storage systems with the valid WWPN(for Fibre Channel) or IQN(for iSCSI) of the node. 

3.2. For Fibre Channel, configure the relevant zoning from the storage to the host.

3.3. For iSCSI, perform the following steps:

3.3.1. Make sure that the login to the iSCSI targets is permanent and remains available after a reboot of the worker node. To do this, verify that the node.startup in the /etc/iscsi/iscsid.conf file is set to automatic. If not, set it as required and then restart the iscsid service `$ service iscsid restart`.

3.3.2. Discover and log into at least two iSCSI targets on the relevant storage systems. (NOTE: Without at least two ports, multipath device will not be created.)

```bash
$ iscsiadm -m discoverydb -t st -p ${STORAGE-SYSTEM-iSCSI-PORT-IP1}:3260 --discover
$ iscsiadm -m node -p ${STORAGE-SYSTEM-iSCSI-PORT-IP1} --login

$ iscsiadm -m discoverydb -t st -p ${STORAGE-SYSTEM-iSCSI-PORT-IP2}:3260 --discover
$ iscsiadm -m node -p ${STORAGE-SYSTEM-iSCSI-PORT-IP2} --login
```

3.3.3. Verify that the login was successful and display all targets that you logged into. The portal value must be the iSCSI target IP address.

```bash
$ iscsiadm -m session --rescan
Rescanning session [sid: 1, target: {storage system IQN},
portal: {STORAGE-SYSTEM-iSCSI-PORT-IP1},{port number}
portal: {STORAGE-SYSTEM-iSCSI-PORT-IP2},{port number}
```

End of worker node setup.




<br/>
<br/>
<br/>


## Installation

### Install the operator



1. Download the manifest from GitHub.
```bash
curl https://raw.githubusercontent.com/IBM/ibm-block-csi-operator/master/deploy/installer/generated/ibm-block-csi-operator.yaml > ibm-block-csi-operator.yaml
```
2. (Optional): If required, update the image fields in the ibm-block-csi-operator.yaml.

3. Install the operator.

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

```bash
$ kubectl apply -f csi.ibm.com_v1_ibmblockcsi_cr.yaml
```

### Verify the driver is running:

```bash
$ kubectl get all -n kube-system  -l csi
NAME                             READY   STATUS    RESTARTS   AGE
pod/ibm-block-csi-controller-0   4/4     Running   0          9m36s
pod/ibm-block-csi-node-jvmvh     3/3     Running   0          9m36s
pod/ibm-block-csi-node-tsppw     3/3     Running   0          9m36s

NAME                                DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/ibm-block-csi-node   2         2         2       2            2           <none>          9m36s

NAME                                        READY   AGE
statefulset.apps/ibm-block-csi-controller   1/1     9m36s
```


## Configuring k8s secret and storage class 
In order to use the driver, create the relevant storage classes and secrets, as needed.

This section describes how to:
 1. Create a storage system secret - to define the storage credential (user and password) and its address.
 2. Configure the k8s storage class - to define the storage system pool name, secret reference, SpaceEfficiency (thin, compressed, or deduplicated) and fstype(xfs\ext4).

#### 1. Create an array secret 
Create a secret file as follows and update the relevant credentials:

```
kind: Secret
apiVersion: v1
metadata:
  name: <VALUE-1>
  namespace: kube-system
type: Opaque
stringData:
  management_address: <VALUE-2,VALUE-3> # Array management addresses
  username: <VALUE-4>                   # Array username
data:
  password: <VALUE-5 base64>            # Array password
```

Apply the secret:

```
$ kubectl apply -f array-secret.yaml
```

#### 2. Create storage classes

Create a storage class yaml file `storageclass-gold.yaml` as follows, with the relevant capabilities, pool and, array secret:

```
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: gold
provisioner: block.csi.ibm.com
parameters:
  #SpaceEfficiency: <VALUE>    # Optional: Values applicable for Storwize are: thin, compressed, or deduplicated
  pool: <VALUE_POOL_NAME>

  csi.storage.k8s.io/provisioner-secret-name: <VALUE_ARRAY_SECRET>
  csi.storage.k8s.io/provisioner-secret-namespace: <VALUE_ARRAY_SECRET_NAMESPACE>
  csi.storage.k8s.io/controller-publish-secret-name: <VALUE_ARRAY_SECRET>
  csi.storage.k8s.io/controller-publish-secret-namespace: <VALUE_ARRAY_SECRET_NAMESPACE>

  csi.storage.k8s.io/fstype: xfs   # Optional: Values ext4/xfs. The default is ext4.
```

Apply the storage class:

```bash
$ kubectl apply -f storageclass-gold.yaml
storageclass.storage.k8s.io/gold created
```
You can now run stateful applications using IBM block storage systems.




<br/>
<br/>
<br/>


## Driver Usage
> **Note**: For further usage details, refer to https://github.com/IBM/ibm-block-csi-driver

<br/>
<br/>
<br/>


## Uninstalling

### 1. Delete the IBMBlockCSI custom resource.
```bash
$ kubectl delete -f csi.ibm.com_v1_ibmblockcsi_cr.yaml
```

### 2. Delete the operator.
```bash
$ kubectl delete -f ibm-block-csi-operator.yaml
```

<br/>
<br/>
<br/>

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

