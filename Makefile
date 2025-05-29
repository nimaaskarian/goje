all: coverage.out tom
SRC = $(shell find -name "*.go" -type f)
WEBGUI = httpd/webgui-preact/dist
WEBGUI_DEPS = $(wildcard httpd/webgui-preact/src/* httpd/webgui-preact/public/* httpd/webgui-preact/index.html)
TW = httpd/webgui-preact/node_modules/tailwindcss httpd/webgui-preact/node_modules/@tailwindcss/vite
coverage.out: $(SRC)
	go test ./... -coverprofile=coverage.out

tom: ${WEBGUI} $(SRC)
	go build

${WEBGUI}: ${TW} $(WEBGUI_DEPS)
	cd httpd/webgui-preact/; npm run build

${TW}:
	cd httpd/webgui-preact/; npm install tailwindcss @tailwindcss/vite

