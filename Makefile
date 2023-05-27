all: mitmproxy

.PHONY: mitmproxy

go_cmd = go.exe
mitmproxy:
	$(go_cmd) build -o go-mitmproxy cmd/go-mitmproxy/*.go

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
