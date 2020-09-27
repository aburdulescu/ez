module github.com/aburdulescu/ez/ezs

go 1.14

require (
	github.com/aburdulescu/ez/chunks v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.4.2
	google.golang.org/protobuf v1.25.0
)

replace github.com/aburdulescu/ez/hash => ../hash

replace github.com/aburdulescu/ez/chunks => ../chunks
