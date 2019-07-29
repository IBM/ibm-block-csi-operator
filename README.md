# ibm-block-csi-driver-operator
The Container Storage Interface (CSI) Driver for IBM block storage systems enables container orchestrators such as Kubernetes to manage the life-cycle of persistent storage.

It is the official operator to deploy and manage the CSI Driver for IBM block storage systems

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

#### Prerequisite

###### Install CSIDriver CRD - optional
Enabling CSIDriver on Kubernetes (more details -> https://kubernetes-csi.github.io/docs/csi-driver-object.html#enabling-csidriver-on-kubernetes)

In Kubernetes v1.13, because the feature was alpha, it was disabled by default. To enable the use of CSIDriver on these versions, do the following:

1. Ensure the feature gate is enabled via the following Kubernetes feature flag: --feature-gates=CSIDriverRegistry=true
   For example on kubeadm installation add the flag inside the /etc/kubernetes/manifests/kube-apiserver.yaml.
2. Either ensure the CSIDriver CRD is automatically installed via the Kubernetes Storage CRD addon OR manually install the CSIDriver CRD on the Kubernetes cluster with the following command:
   ```sh
   #> kubectl create -f https://raw.githubusercontent.com/kubernetes/csi-api/master/pkg/crd/manifests/csidriver.yaml
   ```

If the feature gate was not enabled then CSIDriver for the ibm-block-csi-driver will not be created automatically.

#### 1. Install the CSI driver operator
```sh

#> curl https://raw.githubusercontent.com/IBM/ibm-block-csi-driver-operator/develop/deploy/ibm-block-csi-driver-operator.yaml > ibm-block-csi-driver-operator.yaml 

### Optional: Edit the `ibm-block-csi-driver-operator.yaml` file if you need to change the driver IMAGE URL and the listening port.

#> kubectl apply -f ibm-block-csi-driver-operator.yaml
serviceaccount/ibm-block-csi-controller-sa created
clusterrole.rbac.authorization.k8s.io/ibm-block-csi-external-provisioner-role created
clusterrolebinding.rbac.authorization.k8s.io/ibm-block-csi-external-provisioner-binding created
clusterrole.rbac.authorization.k8s.io/ibm-block-csi-external-attacher-role created
clusterrolebinding.rbac.authorization.k8s.io/ibm-block-csi-external-attacher-binding created
clusterrole.rbac.authorization.k8s.io/ibm-block-csi-cluster-driver-registrar-role created
clusterrolebinding.rbac.authorization.k8s.io/ibm-block-csi-cluster-driver-registrar-binding created
clusterrole.rbac.authorization.k8s.io/ibm-block-csi-external-snapshotter-role created
clusterrolebinding.rbac.authorization.k8s.io/ibm-block-csi-external-snapshotter-binding created
statefulset.apps/ibm-block-csi-controller created
daemonset.apps/ibm-block-csi-node created
```

Verify operator is running (The ibm-block-csi-driver-operator pod should be in Running state):
```sh
#> kubectl get -n kube-system pod --selector=app=ibm-block-csi-controller
NAME                         READY   STATUS    RESTARTS   AGE
ibm-block-csi-controller-0   5/5     Running   0          10m

#> kubectl get -n kube-system pod --selector=app=ibm-block-csi-node
NAME                       READY   STATUS    RESTARTS   AGE
ibm-block-csi-node-xnfgp   3/3     Running   0          10m
ibm-block-csi-node-zgh5h   3/3     Running   0          10m
```

#### 2. Create an IBMBlockCSI custom resource
The operator is running, now you can create an IBMBlockCSI custom resource to install IBM block CSI Driver.
 
Create an IBMBlockCSI file (ibc.yaml) as follow and update the relevant fields:
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

## Un-installation

#### 1. Delete the IBMBlockCSI custom resource
```
#> kubectl delete -f ibc.yaml
```


#### 2. Delete the operator

```sh
#> kubectl delete -f ibm-block-csi-driver-operator.yaml
```

Kubernetes version 1.13 automatically creates the CSIDriver `ibm-block-csi-driver`, but it does not delete it automatically when removing the driver manifest.
So in order to clean up CSIDriver object, run the following commend:
```sh
kubectl delete CSIDriver ibm-block-csi-driver
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

