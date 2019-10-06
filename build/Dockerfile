FROM registry.access.redhat.com/ubi8/ubi-minimal:8.0

ENV OPERATOR=/usr/local/bin/ibm-block-csi-operator \
    USER_UID=1001 \
    USER_NAME=ibm-block-csi-operator

# install operator binary
COPY build/_output/bin/ibm-block-csi-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
