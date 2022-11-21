all: test build

test:
	go test -cover -race ./...

build:
	xk6 build --with xk6-output-prometheus-pushgateway=. --with github.com/mstoykov/xk6-counter@latest

.PHONY: test build