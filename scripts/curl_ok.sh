#! /bin/sh

# Used for a basic HTTP health check
# Avoids output from non-errors and will fail if the HTTP response is unsuccessful

curl --silent --show-error --fail -o /dev/null "$@"
