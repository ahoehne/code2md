appName="code2md"
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
ifeq ($(GOOS),windows)
	fileExtension=".exe"
else
	fileExtension=""
endif
goVersion="1.26"

myUid=1000
myGid=1000

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

build: ## Build for current platform (dist/)
	mkdir -p dist && go test ./... && go mod tidy && go build -v -o "dist/${appName}-${GOOS}-${GOARCH}${fileExtension}"
buildall: ## Build for all platforms
	go test ./... && ./build-all.sh
docker-buildall: ## Build all platforms via Docker
	docker run --rm -it -v "${PWD}":/usr/src/code2md -w /usr/src/code2md "golang:${goVersion}" bash -c "git config --global --add safe.directory /usr/src/code2md && go test ./... && ./build-all.sh && chown -R "${myUid}:${myGid}" dist"

coverage: ## Generate HTML coverage report (cov/)
	if [ -d cov ]; then rm -rf cov; fi
	mkdir -p cov
	go test -coverprofile=cov/coverage.out ./...
	go tool cover -html=cov/coverage.out -o cov/coverage.html

test: ## Run tests with verbose output
	go test ./... -v

install: ## Install to /usr/local/bin (needs sudo)
ifeq ($(GOOS),windows)
	mkdir -p "C:\Program Files\${appName}"
	copy "dist/${appName}-${GOOS}-${GOARCH}${fileExtension}" "C:\Program Files\${appName}\${appName}${fileExtension}"
	setx PATH "%PATH%;C:\Program Files\${appName}"
else
	install -m 755 -o root -g root "dist/${appName}-${GOOS}-${GOARCH}" /usr/local/bin/${appName}
endif

