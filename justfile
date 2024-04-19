# this file is a recipe file for the project

# renovate: datasource=github-releases depName=gotestyourself/gotestsum
gotestsum_version := "v1.11.0"

# Launch a single test using go test in verbose mode
test-verbose testname: install-tests-go-mod
    cd test/src/ && go test -v --timeout=120m -p 1 -run {{testname}}

# Launch a single test using gotestsum
test testname gts_options="": install-tests-go-mod
    cd test/src/ && go run gotest.tools/gotestsum@{{gotestsum_version}} {{gts_options}} -- --timeout=120m -p 1 -run {{testname}}

# Launch the tests in parallel using go test in verbose mode
tests-verbose: install-tests-go-mod
    cd test/src/ && go test -v --timeout=120m -p 1 .

# Launch the tests in parallel using gotestsum
tests gts_options="": install-tests-go-mod
    cd test/src/ && go run gotest.tools/gotestsum@{{gotestsum_version}} {{gts_options}} -- --timeout=120m -p 1 .

# Install go dependencies from test/src/go.mod
install-tests-go-mod:
    cd test/src/ && go mod download

# Install all the tooling
install-tooling: asdf-install

# Install asdf plugins
asdf-plugins:
    #!/bin/sh
    echo "Installing asdf plugins"
    for plugin in $(awk '{print $1}' .tool-versions); do \
      asdf plugin add ${plugin} 2>&1 | (grep "already added" && exit 0); \
    done

    echo "Update all asdf plugins"
    asdf plugin update --all

# Install tools using asdf
asdf-install: asdf-plugins
    asdf install
