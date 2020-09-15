module github.com/aburdulescu/go-ez

go 1.14

require (
	github.com/aburdulescu/go-ez/cli v0.0.0-20200914200009-557db369a291
	github.com/aburdulescu/go-ez/ezt v0.0.0-00010101000000-000000000000
	github.com/dgraph-io/badger/v2 v2.2007.2
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/aburdulescu/go-ez/cli => ./cli

replace github.com/aburdulescu/go-ez/ezt => ./ezt

replace github.com/aburdulescu/go-ez/ezs => ./ezs
