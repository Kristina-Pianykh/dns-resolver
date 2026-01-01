.PHONY: client server

client:
	cd client && go build ./

server:
	cd server && go build ./

query:
	drill -V 5 -4 @192.168.2.223 -p 8085 kristina.pianykh.xyz
