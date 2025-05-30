SRC = $(shell find -name "*.go" -type f)
WEBGUI = httpd/webgui-preact/dist
WEBGUI_DEPS = $(wildcard httpd/webgui-preact/src/* httpd/webgui-preact/public/* httpd/webgui-preact/index.html)
TW = httpd/webgui-preact/node_modules/tailwindcss httpd/webgui-preact/node_modules/@tailwindcss/vite
ALL = coverage.out goje goje.exe

all: ${ALL}

coverage.out: $(SRC)
	go test ./... -coverprofile=coverage.out

goje: ${WEBGUI} $(SRC)
	go build

goje.exe: ${WEBGUI} $(SRC)
	GOOS=windows go build

${WEBGUI}: ${TW} $(WEBGUI_DEPS)
	cd httpd/webgui-preact/; npm run build

${TW}:
	cd httpd/webgui-preact/; npm install tailwindcss @tailwindcss/vite

clean:
	rm ${ALL}
	rm -r ${WEBGUI}
