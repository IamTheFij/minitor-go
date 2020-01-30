ARG REPO=library
FROM multiarch/qemu-user-static:4.2.0-2 as qemu-user-static
# Make sure a dummy x86_64 file exists so that the copy command doesn't error
RUN touch /usr/bin/qemu-x86_64-fake

FROM ${REPO}/alpine:3.10

# Copying all qemu files because amd64 doesn't exist and cannot condional copy
COPY --from=qemu-user-static /usr/bin/qemu-* /usr/bin/

RUN mkdir /app
WORKDIR /app/

# Add common checking tools
RUN apk --no-cache add bash=~5.0 curl=~7.66 jq=~1.6

# Add minitor user for running as non-root
RUN addgroup -S minitor && adduser -S minitor -G minitor

# Copy scripts
COPY ./scripts /app/scripts
RUN chmod -R 755 /app/scripts

# Copy minitor in
ARG ARCH=amd64
COPY ./minitor-linux-${ARCH} ./minitor

# Drop to non-root user
USER minitor

ENTRYPOINT [ "./minitor" ]
