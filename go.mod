module github.com/aburdulescu/ez

go 1.14

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/aburdulescu/ez/chunks v0.0.0-20200927092412-8d7e1d6cc4ec
	github.com/aburdulescu/ez/cli v0.0.0-20200927092412-8d7e1d6cc4ec
	github.com/aburdulescu/ez/ezs v0.0.0-20200927092412-8d7e1d6cc4ec
	github.com/aburdulescu/ez/ezt v0.0.0-20200927092412-8d7e1d6cc4ec
	github.com/aburdulescu/ez/hash v0.0.0-20200927092412-8d7e1d6cc4ec
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20201002202402-0a1ea396d57c // indirect
	google.golang.org/protobuf v1.25.0
)

replace github.com/aburdulescu/ez/hash => ./hash

replace github.com/aburdulescu/ez/cli => ./cli

replace github.com/aburdulescu/ez/chunks => ./chunks

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs
