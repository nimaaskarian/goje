PREFIX:=/usr
SRC = $(shell find -name "*.go" -type f)
WEBGUI = httpd/webgui-preact/dist
WEBGUI_DEPS = $(wildcard httpd/webgui-preact/src/* httpd/webgui-preact/public/* httpd/webgui-preact/index.html)
TW = httpd/webgui-preact/node_modules/tailwindcss httpd/webgui-preact/node_modules/@tailwindcss/vite
ANDROID_NDK_HOME:=/opt/android-sdk/ndk/27.0.12077973
SERVICE_USER_DIR=${DESTDIR}${PREFIX}/lib/systemd/user
SERVICE_SYSTEM_DIR=${DESTDIR}${PREFIX}/lib/systemd/system
SERVICES_SYSTEM=$(wildcard systemd/*@.service)
SERVICES_USER=$(wildcard systemd/*[a-z].service)
SERVICES=$(SERVICES_SYSTEM) $(SERVICES_USER)
BIN_DIR=${DESTDIR}${PREFIX}/bin

all: coverage.out goje $(SERVICES)

coverage.out: $(WEBGUI) $(SRC)
	go test ./... -coverprofile=coverage.out

goje: ${WEBGUI} $(SRC)
	go build -o $@

goje_linux_amd64: ${WEBGUI} $(SRC)
	GOOS=linux go build -o $@

goje_windows_amd64.exe: ${WEBGUI} $(SRC)
	GOOS=windows go build -o $@

goje_android_arm64: ${WEBGUI} $(SRC)
	GOARCH=arm64 CC=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android30-clang\
				 GOOS=android CGO_ENABLED=1 go build -o $@

${WEBGUI}: ${TW} $(WEBGUI_DEPS)
	cd httpd/webgui-preact/; npm run build

${TW}:
	cd httpd/webgui-preact/; npm install tailwindcss @tailwindcss/vite

systemd/%.service: systemd/%.template
	sed 's+BINDIR+${BIN_DIR}+' $< > $@

install: all
	mkdir -p ${BIN_DIR}
	cp -f goje ${BIN_DIR}
	cp -f ${SERVICES_USER} ${SERVICE_USER_DIR}
	cp -f ${SERVICES_SYSTEM} ${SERVICE_SYSTEM_DIR}

uninstall:
	rm ${BIN_DIR}/goje ${SERVICE_FILE}

clean:
	rm ${ALL}
	rm -r ${WEBGUI}
