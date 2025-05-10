all: client server
SRC_SERVER = $(wildcard requests/*.go timer/*.go cmd/server/*.go)
SRC_CLIENT = $(wildcard requests/*.go timer/*.go cmd/client/*.go)
client: $(SRC_CLIENT)
	cd cmd/client && go build
	mv cmd/client/client bin

server: $(SRC_SERVER)
	cd cmd/server && go build
	mv cmd/server/server bin
