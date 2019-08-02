# ibm-block-csi-driver-operator
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

DISCLAIMER: The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.

## Installation

### Prerequisite

#### Install CSIDriver CRD - optional
Enabling CSIDriver on Kubernetes (more details -> https://kubernetes-csi.github.io/docs/csi-driver-object.html#enabling-csidriver-on-kubernetes)

In Kubernetes v1.13, because the feature was alpha, it was disabled by default. To enable the use of CSIDriver on these versions, do the following:

1. Ensure the feature gate is enabled via the following Kubernetes feature flag: --feature-gates=CSIDriverRegistry=true
   For example on kubeadm installation add the flag inside the /etc/kubernetes/manifests/kube-apiserver.yaml.
2. Either ensure the CSIDriver CRD is automatically installed via the Kubernetes Storage CRD addon OR manually install the CSIDriver CRD on the Kubernetes cluster with the following command:
   ```sh
   #> kubectl create -f https://raw.githubusercontent.com/kubernetes/csi-api/master/pkg/crd/manifests/csidriver.yaml
   ```

If the feature gate was not enabled then CSIDriver for the ibm-block-csi-driver will not be created automatically.

### Install the CSI driver operator

#### Install with helm
```sh

#> helm repo add artifactory https://stg-artifactory.haifa.ibm.com/artifactory/chart-repo
#> helm install --name ibm-block-csi-driver-operator --namespace kube-system artifactory/ibm-block-csi-driver-operator

```
#### Install with yaml
```sh

#> kubectl apply -f deploy/csi_driver.yaml  (install csi_driver.yaml only if you are using Kubernetes v.1.14+)
#> kubectl apply -f deploy/ibm-block-csi-driver-operator.yaml

```

### Verify operator is running (The ibm-block-csi-driver-operator pod should be in Running state):
```sh
#> kubectl get pod -l app.kubernetes.io/name=ibm-block-csi-driver-operator -n kube-system
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
```
#> kubectl apply -f ibc.yaml
```

## Uninstallation

### 1. Delete the IBMBlockCSI custom resource
```
#> kubectl delete -f ibc.yaml
```


### 2. Delete the operator

#### Delete with helm
```sh

#> helm delete --purge ibm-block-csi-driver-operator

```
#### Delete with yaml
```sh

#> kubectl delete CSIDriver ibm-block-csi-driver
#> kubectl delete -f deploy/ibm-block-csi-driver-operator.yaml

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

