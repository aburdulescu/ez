module github.com/aburdulescu/go-ez/stream

go 1.14

replace github.com/aburdulescu/go-ez/stream/rpc => ./rpc

require (
	github.com/aburdulescu/go-ez/stream/rpc v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.25.0
)
