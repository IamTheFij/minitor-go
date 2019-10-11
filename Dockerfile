ARG REPO=library
FROM ${REPO}/busybox:latest
WORKDIR /root/

ARG ARCH=amd64
COPY ./minitor-go ./minitor

ENTRYPOINT [ "./minitor" ]
