# Contributing guidelines

## Prerequisites
- [go](https://golang.org/dl/) version v1.13+.
- [docker](https://docs.docker.com/install/) version 17.03+.
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.14.1+.
- [kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) v2.0.1+

## How to run unit tests locally
1. Make sure kubebuilder is installed.
2. Run `make test` in the project root directory.

## Developing rules
1.  If CRDs in `deploy/crds` are updated, you need to update the file with the same name in `deploy/olm-catalog/ibm-block-csi-operator` accordingly.

2.  If `role.yaml`, `role_binding.yaml` or `operator.yaml` is in `deploy` are updated, you need to update the ClusterServiceVersion(CSV) file in `deploy/olm-catalog/ibm-block-csi-operator` accordingly.

3. If README.md is updated, you need to update the ClusterServiceVersion(CSV) file in `deploy/olm-catalog/ibm-block-csi-operator` accordingly.

4. Make sure to run `make update` or `make test` before you finish a commit.
