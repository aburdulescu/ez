module github.com/aburdulescu/ez/chunks

go 1.14

replace github.com/aburdulescu/ez/hash => ../hash

require (
	github.com/aburdulescu/ez/hash v0.0.0-20201220110038-a9dcb59599cf
	github.com/zeebo/xxh3 v0.9.0 // indirect
	golang.org/x/sys v0.0.0-20201221093633-bc327ba9c2f0 // indirect
)
