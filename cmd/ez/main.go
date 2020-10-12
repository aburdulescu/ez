package main

import (
	"log"
	"os"
	// "github.com/pkg/profile"
)

func main() {
	// defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()
	// defer profile.Start(profile.ProfilePath("."), profile.MemProfile, profile.MemProfileRate(1)).Stop()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
