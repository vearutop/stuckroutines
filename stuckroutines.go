package main

import (
	"flag"
	"fmt"

	"github.com/bool64/dev/version"
	"github.com/vearutop/stuckroutines/stuckroutines"
)

func main() {
	f := stuckroutines.Flags{}
	f.Register()

	ver := flag.Bool("version", false, "Print version")

	flag.Parse()

	if *ver {
		fmt.Println(version.Info().Version)

		return
	}

	stuckroutines.Run(f)
}
