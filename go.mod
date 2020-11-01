module github.com/aburdulescu/ez

go 1.14

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/aburdulescu/ez/chunks v0.0.0-20201025134226-187771104c07
	github.com/aburdulescu/ez/ezt v0.0.0-20201025134226-187771104c07
	github.com/aburdulescu/ez/hash v0.0.0-20201025134226-187771104c07
	github.com/aburdulescu/ez/swp v0.0.0-00010101000000-000000000000
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/fatih/color v1.10.0 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.1.1
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/aburdulescu/ez/hash => ./hash

replace github.com/aburdulescu/ez/cli => ./cli

replace github.com/aburdulescu/ez/chunks => ./chunks

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs

replace github.com/aburdulescu/ez/swp => ./swp
