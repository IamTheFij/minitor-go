#! /bin/bash
set -e

#################
# docker_check.sh
#
# Checks the most recent state exit code of a Docker container
#################

# Docker host will default to a socket
# To override, export DOCKER_HOST to a new hostname
DOCKER_HOST="${DOCKER_HOST:=socket}"
container_name="$1"
num_log_lines="$2"

# Curls Docker either using a socket or URL
function curl_docker {
    local path="$1"
    if [ "$DOCKER_HOST" == "socket" ]; then
        curl --unix-socket /var/run/docker.sock "http://localhost/$path" 2>/dev/null
    else
        curl "http://${DOCKER_HOST}/$path" 2>/dev/null
    fi
}

# Returns caintainer ID for a given container name
function get_container_id {
    local container_name="$1"
    curl_docker 'containers/json?all=1' \
        | jq -r ".[] | {Id, Name: .Names[]} | select(.Name == \"/${container_name}\") | .Id"
}

# Returns container JSON
function inspect_container {
    local container_id="$1"
    curl_docker "containers/$container_id/json"
}

# Gets some lines from docker log
function get_logs {
    container_id="$1"
    num_lines="$2"
    curl_docker "containers/$container_id/logs?stdout=1&stderr=1" | tail -n "$num_lines"
}

if [ -z "$container_name" ]; then
    echo "Usage: $0 container_name [num_log_lines]"
    echo "Will exit with the last status code of continer with provided name"
    exit 1
fi

container_id=$(get_container_id "$container_name")
if [ -z "$container_id" ]; then
    echo "ERROR: Could not find container with name: $container_name"
    exit 1
fi
exit_code=$(inspect_container "$container_id" | jq -r .State.ExitCode)

if [ -n "$num_log_lines" ]; then
    get_logs "$container_id" "$num_log_lines"
fi

exit "$exit_code"
