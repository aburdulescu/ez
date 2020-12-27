module github.com/aburdulescu/ez

go 1.14

require (
	github.com/aburdulescu/ez/cmn v0.0.0-20201227061956-1d7845970953
	github.com/aburdulescu/ez/ezt v0.0.0-20201227061956-1d7845970953
	github.com/aburdulescu/ez/swp v0.0.0-20201227061956-1d7845970953
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/fatih/color v1.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/spf13/cobra v1.1.1
	go.etcd.io/bbolt v1.3.5
	golang.org/x/sys v0.0.0-20201223074533-0d417f636930 // indirect
)

replace github.com/aburdulescu/ez/cmn => ./cmn

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs

replace github.com/aburdulescu/ez/swp => ./swp
