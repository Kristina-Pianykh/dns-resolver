.PHONY: client server

build:
	go build ./

test:
	go test -v ./...

query:
	drill -V 5 -4 @192.168.2.223 -p 8085 kristina.pianykh.xyz
