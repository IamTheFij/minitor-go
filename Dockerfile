ARG REPO=library
FROM ${REPO}/alpine:3.10
RUN mkdir /app
WORKDIR /app/

# Copy minitor in
ARG ARCH=amd64
COPY ./minitor-go ./minitor

# Add common checking tools
RUN apk --no-cache add bash=~5.0 curl=~7.66 jq=~1.6

# Add minitor user for running as non-root
RUN addgroup -S minitor && adduser -S minitor -G minitor

# Copy scripts
COPY ./scripts /app/scripts
RUN chown -R minitor:minitor /app
RUN chmod -R 755 /app/scripts

# Drop to non-root user
USER minitor

ENTRYPOINT [ "./minitor" ]
