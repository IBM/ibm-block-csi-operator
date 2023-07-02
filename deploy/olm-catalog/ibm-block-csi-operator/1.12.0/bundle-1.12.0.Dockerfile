FROM scratch

LABEL operators.operatorframework.io.bundle.channel.default.v1=stable
LABEL operators.operatorframework.io.bundle.channels.v1=stable
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=ibm-block-csi-operator

COPY deploy/olm-catalog/ibm-block-csi-operator/1.12.0/manifests /manifests/
COPY deploy/olm-catalog/ibm-block-csi-operator/1.12.0/metadata /metadata/
LABEL com.redhat.openshift.versions="v4.11-v4.13"
LABEL com.redhat.delivery.operator.bundle=true
