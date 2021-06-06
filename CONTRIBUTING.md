# Contributing guidelines

## Prerequisites
- [go](https://golang.org/dl/) version v1.13+.
- [docker](https://docs.docker.com/install/) version 17.03+.
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.14.1+.
- [kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) v2.0.1+

## How to run unit tests locally
1. Make sure kubebuilder is installed.
2. Run `make test` in the project root directory.

## Developing guidelines
- Be sure to run `make update` or `make test` before you finish a commit.
- If CRDs in `deploy/crds` are updated, the same file names located in `deploy/olm-catalog/ibm-block-csi-operator` must be updated accordingly.
- If `role.yaml`, `role_binding.yaml`, or `operator.yaml` in `deploy` are updated, the ClusterServiceVersion(CSV) file in `deploy/olm-catalog/ibm-block-csi-operator` must be updated accordingly.
- If `README.md` is updated, ClusterServiceVersion(CSV) file in `deploy/olm-catalog/ibm-block-csi-operator` must be updated accordingly.
- Run `operator-sdk add` to add a new API or controller, for more details, please refer to https://github.com/operator-framework/operator-sdk.
- Run `operator-sdk generate k8s` and `operator-sdk generate crds` after you change something in `pkg/apis`.

## Package the Operator
This repository makes use of the [Operator Framework](https://github.com/operator-framework) and its packaging concept for Operators. Make sure you read the following guides before packaging the operator and uploading to OperatorHub.
- https://github.com/operator-framework/community-operators/blob/master/docs/contributing.md
- https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/design/building-your-csv.md
- https://github.com/operator-framework/community-operators/blob/master/docs/required-fields.md
