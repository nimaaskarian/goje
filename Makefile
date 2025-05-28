all: coverage.out webgui tom
SRC = $(shell find -name "*.go" -type f)
WEBGUI = $(wildcard httpd/webgui-preact/*)
coverage.out: $(SRC)
	go test ./... -coverprofile=coverage.out

tom: $(SRC)
	go build

webgui: $(WEBGUI)
	cd httpd/webgui-preact/; npm run build
