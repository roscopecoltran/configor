
default: all
all: build test

build:
	cd ./example && go build -o configore-test main.go

test:
	cd ./example && ./configore-test 