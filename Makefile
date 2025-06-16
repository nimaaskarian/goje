SRC = $(shell find -name "*.go" -type f)
WEBGUI = httpd/webgui-preact/dist
WEBGUI_DEPS = $(wildcard httpd/webgui-preact/src/* httpd/webgui-preact/public/* httpd/webgui-preact/index.html)
TW = httpd/webgui-preact/node_modules/tailwindcss httpd/webgui-preact/node_modules/@tailwindcss/vite
ANDROID_NDK_HOME:=/opt/android-sdk/ndk/27.0.12077973

all: TODO.md coverage.out goje_linux_amd64

coverage.out: $(WEBGUI) $(SRC)
	go test ./... -coverprofile=coverage.out

goje_linux_amd64: ${WEBGUI} $(SRC)
	go build -o $@

TODO.md: tasks.yaml
	ydo md > TODO.md

goje_windows_amd64.exe: ${WEBGUI} $(SRC)
	GOOS=windows go build -o $@

goje_android_arm64: ${WEBGUI} $(SRC)
	GOARCH=arm64 CC=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android30-clang\
				 GOOS=android CGO_ENABLED=1 go build -o $@

${WEBGUI}: ${TW} $(WEBGUI_DEPS)
	cd httpd/webgui-preact/; npm run build

${TW}:
	cd httpd/webgui-preact/; npm install tailwindcss @tailwindcss/vite

install: all
	mkdir -p ${DESTDIR}${PREFIX}/bin
	cp -f dwm ${DESTDIR}${PREFIX}/bin

clean:
	rm ${ALL}
	rm -r ${WEBGUI}
