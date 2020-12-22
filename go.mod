module github.com/aburdulescu/ez

go 1.14

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/aburdulescu/ez/chunks v0.0.0-20201220110038-a9dcb59599cf
	github.com/aburdulescu/ez/ezt v0.0.0-20201220110038-a9dcb59599cf
	github.com/aburdulescu/ez/hash v0.0.0-20201220110038-a9dcb59599cf
	github.com/aburdulescu/ez/swp v0.0.0-20201220110038-a9dcb59599cf
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
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20201216054612-986b41b23924 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/aburdulescu/ez/hash => ./hash

replace github.com/aburdulescu/ez/cli => ./cli

replace github.com/aburdulescu/ez/chunks => ./chunks

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs

replace github.com/aburdulescu/ez/swp => ./swp
