all: test build
.PHONY: all

test:
	go test -cover -race ./...

build:
	xk6 build --with github.com/martymarron/xk6-output-prometheus-pushgateway@latest=.

run:
	xk6 run \
	-o output-prometheus-pushgateway \
	--iterations 100 \
	--vus 20 \
	--verbose \
	./script.js
