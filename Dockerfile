ARG REPO=library
FROM ${REPO}/alpine:3.17

RUN mkdir /app
WORKDIR /app/

# Add common checking tools
RUN apk --no-cache add bash=~5 curl=~8 jq=~1.6 bind-tools~=9

# Add minitor user for running as non-root
RUN addgroup -S minitor && adduser -S minitor -G minitor

# Copy scripts
COPY ./scripts /app/scripts
RUN chmod -R 755 /app/scripts

# Copy minitor in
ARG ARCH=amd64
COPY ./dist/minitor-linux-${ARCH} ./minitor

# Drop to non-root user
USER minitor

ENTRYPOINT [ "./minitor" ]
