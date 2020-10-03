module github.com/aburdulescu/ez/ezs

go 1.14

require (
	github.com/aburdulescu/ez/chunks v0.0.0-20200927092412-8d7e1d6cc4ec
	github.com/golang/protobuf v1.4.2
	google.golang.org/protobuf v1.25.0
)

replace github.com/aburdulescu/ez/hash => ../hash

replace github.com/aburdulescu/ez/chunks => ../chunks
