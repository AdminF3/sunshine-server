export GO111MODULE=on

.PHONY: all build

ENV ?= dev

all: build migrate

build: dep ## Build binary file
	go install -ldflags "-X stageai.tech/sunshine/sunshine.version=$$(date +%Y%m%d)-$$(git rev-parse --short HEAD)" stageai.tech/sunshine/sunshine/cmd/...

dep: ## Install dependencies.
	git config --global url.git@gitlab.com:.insteadOf https://gitlab.com
	which goose > /dev/null || go get github.com/pressly/goose/cmd/goose@v2.6.0
	which staticcheck > /dev/null || go get honnef.co/go/tools/cmd/staticcheck@2020.1.4
	which dataloaden > /dev/null || go get github.com/vektah/dataloaden@v0.3.0
	go generate ./...
	go get ./...

migrate: ## Execute database migrations
	SUNSHINE_ENV=$(ENV) sunshine migrate

## Diff command for `go fmt` needs bash.
test: SHELL:=/bin/bash
test: ## Run tests.
	SUNSHINE_ENV=$(ENV) go test -coverprofile=coverage.out -covermode=count -coverpkg=./... -timeout 30m ./...
	@diff -u <(echo -n) <(go fmt ./...) || echo "Run go fmt -w on save!"
	go vet ./...
	staticcheck ./...

coverage.html: ## Generate HTML code coverage report
	go tool cover -html=coverage.out -o coverage.html

coverage.xml: ## Generate XML code coverage report
	gocov convert coverage.out | gocov-xml > coverage.xml

clean: ## Removes auto-generated fiels and project binary.
	rm -f coverage.xml
	rm -f coverage.html
	rm -f coverage.out
	rm -f $(GOPATH)/bin/sunshine
	rm -f openapi.json
	find . -name "*_gen.go" | xargs -r rm


help: ## Display help screen.
	@grep -h -E '^[\.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
