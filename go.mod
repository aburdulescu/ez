module github.com/aburdulescu/ez

go 1.16

require (
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/aburdulescu/ez/cadet v0.0.0-20201229150525-18426b6b5b81
	github.com/aburdulescu/ez/cmn v0.0.0-20201229150525-18426b6b5b81
	github.com/aburdulescu/ez/ezt v0.0.0-20201229150525-18426b6b5b81
	github.com/aburdulescu/ez/swp v0.0.0-20201229150525-18426b6b5b81
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/fatih/color v1.11.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	go.etcd.io/bbolt v1.3.5
	golang.org/x/sys v0.0.0-20210521090106-6ca3eb03dfc2 // indirect
)

replace github.com/aburdulescu/ez/cmn => ./cmn

replace github.com/aburdulescu/ez/ezt => ./ezt

replace github.com/aburdulescu/ez/ezs => ./ezs

replace github.com/aburdulescu/ez/swp => ./swp

replace github.com/aburdulescu/ez/cadet => ./cadet
