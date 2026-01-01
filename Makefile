.PHONY: client server

client:
	cd client && go build ./

server:
	cd server && go build ./
