FROM alpine:3.23

RUN mkdir /app
WORKDIR /app/

# Add common checking tools
# hadolint ignore=DL3018
RUN apk --no-cache add bash=~5 curl=~8 jq=~1 bind-tools=~9 tzdata

# Add minitor user for running as non-root
RUN addgroup -S minitor && adduser -S minitor -G minitor

# Copy scripts
COPY ./scripts /app/scripts
RUN chmod -R 755 /app/scripts

# Copy minitor in
ARG TARGETOS
ARG TARGETARCH
COPY ./dist/minitor-${TARGETOS}-${TARGETARCH} ./minitor

# Drop to non-root user
USER minitor

ENTRYPOINT [ "./minitor" ]
