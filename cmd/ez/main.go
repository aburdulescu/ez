package main

import (
	"fmt"
	"log"
	"os"
	// "github.com/pkg/profile"
)

func main() {
	// defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()
	// defer profile.Start(profile.ProfilePath("."), profile.MemProfile, profile.MemProfileRate(1)).Stop()
	// defer profile.Start(profile.ProfilePath("."), profile.TraceProfile).Stop()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)

	if err := root.AddCommand(commands...); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	root.Execute()
}
