# Build the manager binary
FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY pkg/ pkg/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.7-1107

ARG VERSION=1.11.0
ARG BUILD_NUMBER=0

###Required Labels
LABEL name="IBM block CSI operator" \
    vendor="IBM" \
    version=$VERSION \
    release=$BUILD_NUMBER \
    summary="Manage objects in kubernetes and openshift" \
    description="The IBM block CSI operator enables container orchestrators to use objects and to manage them in their storage." \
    io.k8s.display-name="IBM block CSI operator" \
    io.k8s.description="The IBM block CSI operator enables container orchestrators to use objects and to manage them in their storage." \
    io.openshift.tags=ibm,csi,volume-group-operator

COPY ./LICENSE /licenses/

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
