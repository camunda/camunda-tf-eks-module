#!/bin/bash

# Run a single Docker container to handle the README.md updates
docker run --rm \
    -v "$PWD":/workspace \
    -w /workspace \
    node:22 \
    bash -c '
        npm install -g action-docs
        find .github/actions -name "*.yml" -o -name "*.yaml" | while read -r action_file; do
            action_dir=$(dirname "$action_file")
            echo "Updating README.md in $action_dir"
            rm -f "$action_dir/README.md"
            action-docs -t 1 --no-banner -n -s "$action_file" > "$action_dir/README.md.tmp"
            # Ensure that only a single empty line is left at the end of the file
            sed -e :a -e "/^\n*\$/{\$d;N;};/\n\$/ba" "$action_dir/README.md.tmp" > "$action_dir/README.md"
            rm -f "$action_dir/README.md.tmp"
        done
    '
