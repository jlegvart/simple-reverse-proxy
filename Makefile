BINARY_NAME=simple-proxy

build:
	go build -o bin/${BINARY_NAME} server.go

run: build
	bin/./${BINARY_NAME}