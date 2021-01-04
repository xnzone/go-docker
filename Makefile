# GOPATH:=$(shell go env GOPATH)

#需要修改部分
BINARY=go-docker
MAIN_FILE=main.go main_command.go run.go

#不要修改下列脚本
.PHONY: build build_local build_64linux yconf docker install clean version v cmp_dep install-go-xray

build: 
	go fmt ./...
	go vet ./...
	go build -o bin/${BINARY} ${MAIN_FILE}

build_local:
	go fmt ./...
	go vet ./...
	go build -o bin/${BINARY} ${MAIN_FILE}


build_64linux:
	go fmt ./...
	go vet ./...
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/${BINARY} ${MAIN_FILE}


test:
	go test -v ./... -cover

docker:
	docker build . -t ${BINARY}:latest

install:
	go install ./...

version v:
	@if [ -f bin/${BINARY} ] ; then bin/${BINARY} -buildinfo; fi

clean:
	@if [ -f bin/${BINARY} ] ; then rm bin/${BINARY} ; fi

