.PHONY: client server

build:
	go build ./

test:
	go test -race -v ./...

query:
	dig @192.168.2.223 -p 8085 kristina.pianykh.xyz +noedns +qr
