#!/bin/bash

set -o pipefail

# due to an open bug in the node image https://github.com/nodejs/docker-node/issues/740
# we can't map the user and group at the docker level, therefore we chown the files
USER_ID=$(id -u)
GROUP_ID=$(id -g)

# Run a single Docker container to handle the README.md updates
docker run --rm \
    -e USER_ID="$USER_ID" \
    -e GROUP_ID="$GROUP_ID" \
    -v "$PWD":/workspace \
    -w /workspace \
    node:22 \
    bash -c '
        set -euxo pipefail
        npm install -g action-docs
        find .github/actions -name "*.yml" -o -name "*.yaml" | while read -r action_file; do
            action_dir=$(dirname "$action_file")
            echo "Updating README.md in $action_dir"
            rm -f "$action_dir/README.md"
            action-docs -t 1 --no-banner -n -s "$action_file" > "$action_dir/README.md.tmp"
            # Ensure that only a single empty line is left at the end of the file
            sed -e :a -e "/^\n*\$/{\$d;N;};/\n\$/ba" "$action_dir/README.md.tmp" > "$action_dir/README.md"
            chown "$USER_ID:$GROUP_ID" "$action_dir/README.md"
            rm -f "$action_dir/README.md.tmp"
        done
    '

DOCKER_EXIT_CODE=$?

if [ $DOCKER_EXIT_CODE -ne 0 ]; then
    echo "Docker action readme generation command failed with exit code $DOCKER_EXIT_CODE, please use verbose mode"
    exit 1
fi
