module github.com/aburdulescu/go-ez

go 1.14

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/aburdulescu/go-ez/chunks v0.0.0-00010101000000-000000000000
	github.com/aburdulescu/go-ez/cli v0.0.0-20200915120451-e2971c4446ad
	github.com/aburdulescu/go-ez/ezs v0.0.0-00010101000000-000000000000
	github.com/aburdulescu/go-ez/ezt v0.0.0-20200915120451-e2971c4446ad
	github.com/aburdulescu/go-ez/hash v0.0.0-00010101000000-000000000000
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20200904194848-62affa334b73 // indirect
	golang.org/x/sys v0.0.0-20200915084602-288bc346aa39 // indirect
	google.golang.org/protobuf v1.25.0
)

replace github.com/aburdulescu/go-ez/cli => ./cli

replace github.com/aburdulescu/go-ez/ezt => ./ezt

replace github.com/aburdulescu/go-ez/ezs => ./ezs

replace github.com/aburdulescu/go-ez/chunks => ./chunks

replace github.com/aburdulescu/go-ez/hash => ./hash
