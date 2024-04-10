#!make
.DEFAULT_GOAL= help

.PHONY : init
init: ## Initialize tests
	@exit 0

.PHONY : test
test: init ## Run tests
	go mod download
	go test -v -timeout 120m

# see https://suva.sh/posts/well-documented-makefiles/
.PHONY : help
help: ## Show this help prompt.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
