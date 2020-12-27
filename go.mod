module github.com/aburdulescu/ez

go 1.14

require (
	github.com/aburdulescu/ez/cadet v0.0.0-00010101000000-000000000000
	github.com/aburdulescu/ez/cmn v0.0.0-20201227104937-05ed050507c2
	github.com/aburdulescu/ez/ezt v0.0.0-20201227104937-05ed050507c2
	github.com/aburdulescu/ez/swp v0.0.0-20201227104937-05ed050507c2
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/fatih/color v1.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/mattn/go-runewidth v0.0.9 // indirect
	go.etcd.io/bbolt v1.3.5
	golang.org/x/sys v0.0.0-20201223074533-0d417f636930 // indirect
)

replace github.com/aburdulescu/ez/cmn => ./cmn

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs

replace github.com/aburdulescu/ez/swp => ./swp

replace github.com/aburdulescu/ez/cadet => ./cadet
