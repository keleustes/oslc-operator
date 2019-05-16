FROM alpine:3.8

ENV OPERATOR=/usr/local/bin/openstacklcm-operator \
    USER_UID=1001 \
    USER_NAME=openstacklcm-operator

# install operator binary
COPY build/_output/bin/openstacklcm-operator ${OPERATOR}

COPY build/bin /usr/local/bin

RUN  /usr/local/bin/user_setup

COPY helm-charts /opt/openstacklcm-operator/helm-charts/

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
