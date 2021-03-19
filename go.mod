module github.com/aburdulescu/ez

go 1.14

require (
	github.com/aburdulescu/ez/cadet v0.0.0-20201229150525-18426b6b5b81
	github.com/aburdulescu/ez/cmn v0.0.0-20201229150525-18426b6b5b81
	github.com/aburdulescu/ez/ezt v0.0.0-20201229150525-18426b6b5b81
	github.com/aburdulescu/ez/swp v0.0.0-20201229150525-18426b6b5b81
	github.com/cheggaaa/pb/v3 v3.0.6
	github.com/fatih/color v1.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	go.etcd.io/bbolt v1.3.5
	golang.org/x/sys v0.0.0-20210315160823-c6e025ad8005 // indirect
)

replace github.com/aburdulescu/ez/cmn => ./cmn

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs

replace github.com/aburdulescu/ez/swp => ./swp

replace github.com/aburdulescu/ez/cadet => ./cadet
