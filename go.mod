module github.com/pplorins/go-mitmproxy

go 1.18

require (
	github.com/andybalholm/brotli v1.0.4
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/gorilla/websocket v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.0.4
	github.com/samber/lo v1.37.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/tidwall/match v1.1.1
	gitlab.com/pplorins/wechat-official-accounts v0.0.0-20230626183203-1af5fd1e966c
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	golang.org/x/crypto v0.10.0 // indirect
	golang.org/x/exp v0.0.0-20220303212507-bbda1eaf7a17 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/term v0.9.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

//for mac only
replace gitlab.com/pplorins/wechat-official-accounts => /Users/arthur/git/wechat-official-accounts
