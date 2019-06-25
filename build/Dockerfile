FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/ibm-block-csi-driver-operator \
    USER_UID=1001 \
    USER_NAME=ibm-block-csi-driver-operator

# install operator binary
COPY build/_output/bin/ibm-block-csi-driver-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
