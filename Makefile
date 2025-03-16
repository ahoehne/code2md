appName="code2md"
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
ifeq ($(GOOS),windows)
	fileExtension=".exe"
else
	fileExtension=""
endif
goVersion="1.24"

myUid=1000
myGid=1000

coverage:
	if [ -d cov ]; then rm -rf cov; fi
	mkdir -p cov
	go test -coverprofile=cov/coverage.out ./...
	go tool cover -html=cov/coverage.out -o cov/coverage.html

test:
	go test ./... -v

build:
	mkdir -p dist && go test ./... && go mod tidy && go build -v -o "dist/${appName}-${GOOS}-${GOARCH}${fileExtension}"

buildall:
	go test ./... && ./build-all.sh

docker-buildall:
	docker run --rm -it -v "${PWD}":/usr/src/code2md -w /usr/src/code2md "golang:${goVersion}" bash -c "git config --global --add safe.directory /usr/src/code2md && go test ./... && ./build-all.sh && chown -R "${myUid}:${myGid}" dist"

install:
ifeq ($(GOOS),windows)
	mkdir -p "C:\Program Files\${appName}"
	copy "dist/${appName}-${GOOS}-${GOARCH}${fileExtension}" "C:\Program Files\${appName}\${appName}${fileExtension}"
	setx PATH "%PATH%;C:\Program Files\${appName}"
else
	install -m 755 -o root -g root "dist/${appName}-${GOOS}-${GOARCH}" /usr/bin/${appName}
endif

