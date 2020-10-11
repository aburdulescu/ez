module github.com/aburdulescu/ez/ezs

go 1.14

require (
	github.com/aburdulescu/ez/chunks v0.0.0-20201011111155-0513dc0f5dc4
	github.com/golang/protobuf v1.4.2
	google.golang.org/protobuf v1.25.0
)

replace github.com/aburdulescu/ez/hash => ../hash

replace github.com/aburdulescu/ez/chunks => ../chunks
