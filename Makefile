all: mitmproxy-linux

.PHONY: mitmproxy

go_cmd = go
mitmproxy:
	$(go_cmd) build -o go-mitmproxy cmd/go-mitmproxy/*.go

mitmproxy-linux:
	GOOS=linux GOARCH=amd64 $(go_cmd) build -o go-mitmproxy-linux cmd/go-mitmproxy/*.go

.PHONY: dummycert
dummycert:
	$(go_cmd) build -o dummycert cmd/dummycert/main.go

.PHONY: clean
clean:
	rm -f go-mitmproxy dummycert

.PHONY: test
test:
	$(go_cmd) test ./... -v

.PHONY: dev
dev:
	$(go_cmd) run $(shell ls cmd/go-mitmproxy/*.go | grep -v _test.go)
