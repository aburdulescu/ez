module github.com/aburdulescu/ez

go 1.14

require (
	github.com/aburdulescu/ez/chunks v0.0.0-20200924203214-ec388fee5085
	github.com/aburdulescu/ez/cli v0.0.0-20200924203214-ec388fee5085
	github.com/aburdulescu/ez/ezs v0.0.0-20200924203214-ec388fee5085
	github.com/aburdulescu/ez/ezt v0.0.0-20200924203214-ec388fee5085
	github.com/aburdulescu/ez/hash v0.0.0-20200924203214-ec388fee5085
	github.com/dgraph-io/badger/v2 v2.2007.2
	golang.org/x/sys v0.0.0-20200923182605-d9f96fdee20d // indirect
	google.golang.org/protobuf v1.25.0
)

replace github.com/aburdulescu/ez/hash => ./hash

replace github.com/aburdulescu/ez/cli => ./cli

replace github.com/aburdulescu/ez/chunks => ./chunks

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs
