.DEFAULT_GOAL := build

TARGET := courier

build:
	go build -o ${TARGET} cmd/main.go && mv ${TARGET} cmd/build/
