
default: all
all: build test

build:
	@echo ""
	@echo "build"
	@cd ./example && go build -o configor-test main.go

test:
	@echo ""
	@echo "test"
	@cd ./example && ./configor-test 