
default: all
all: build test

build:
	cd ./example && go build -o configor-test main.go

test:
	cd ./example && ./configor-test 