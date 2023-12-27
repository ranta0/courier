.DEFAULT_GOAL := build

TARGET := courier
FLAGS := -s -w
GOOS=linux
GOARCH=amd64

build:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="${FLAGS}" -o ${TARGET} cmd/main.go && mv ${TARGET} cmd/build/
