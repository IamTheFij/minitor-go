FROM ${REPO}/alpine:latest
RUN mkdir /app
WORKDIR /app/

# Copy minitor in
ARG ARCH=amd64
COPY ./minitor-go ./minitor

# Add common checking tools
RUN apk --no-cache add bash==4.4.19-r1 curl==7.64.0-r3 jq==1.6-r0

# Add minitor user for running as non-root
RUN addgroup -S minitor && adduser -S minitor -G minitor

# Copy scripts
COPY ./scripts /app/scripts
RUN chmod -R 755 /app/scripts

# Drop to non-root user
USER minitor

ENTRYPOINT [ "./minitor" ]
