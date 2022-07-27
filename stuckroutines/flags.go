package stuckroutines

import (
	"flag"
	"fmt"
	"time"
)

type Flags struct {
	URL           string
	Iterations    int
	Delay         time.Duration
	NoGroup       bool
	SortTrace     bool
	KeepTemporary bool
	MinCount      int
}

func (f *Flags) Register() {
	flag.StringVar(&f.URL, "url", "", "Full URL to /debug/pprof/goroutine?debug=2")
	flag.IntVar(&f.Iterations, "iterations", 2, "How many reports to collect to find persisting routines")
	flag.DurationVar(&f.Delay, "delay", 5*time.Second, "Delay between report collections")
	flag.BoolVar(&f.NoGroup, "no-group", false, "Do not group goroutines by stack trace")
	flag.BoolVar(&f.SortTrace, "sort-trace", false, "Sort by trace instead of count of goroutines")
	flag.BoolVar(&f.KeepTemporary, "keep-temp", false, "Keep temporary goroutines.")
	flag.IntVar(&f.MinCount, "min-count", 10, "Filter traces with few goroutines")

	usage := flag.CommandLine.Usage
	flag.CommandLine.Usage = func() {
		fmt.Println("Stuckroutines requires either a URL or a list of files obtained from /pprof/goroutine?debug=2")
		fmt.Println("Usage: stuckroutines [options] [...report files]")

		usage()
	}
}
