all: client server
client:
		@cd src/client && go build -o ../../bin/chat_client
server:
		@cd src/server && go build -o ../../bin/chat_server
