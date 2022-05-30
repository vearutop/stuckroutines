package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bool64/dev/version"
)

func main() {
	url := flag.String("url", "", "Full URL to /debug/pprof/goroutine?debug=2")
	n := flag.Int("iterations", 2, "How many reports to collect to find persisting routines")
	delay := flag.Duration("delay", 5*time.Second, "Delay between report collections")
	noGroup := flag.Bool("no-group", false, "Do not group goroutines by stack trace")
	ver := flag.Bool("version", false, "Print version")

	usage := flag.CommandLine.Usage
	flag.CommandLine.Usage = func() {
		fmt.Println("Stuckroutines requires either a URL or a list of files obtained from /pprof/goroutine?debug=2")
		fmt.Println("Usage: stuckroutines [options] [...report files]")

		usage()
	}

	flag.Parse()

	if *ver {
		fmt.Println(version.Info().Version)

		return
	}

	if *url == "" && flag.NArg() == 0 {
		flag.Usage()

		return
	}

	r := res{
		result:  make(map[string]goroutine),
		noGroup: *noGroup,
	}

	if *url != "" {
		r.fetch(*n, *url, *delay)
	}

	for _, fn := range flag.Args() {
		r.load(fn)
	}

	r.count()

	println(r.persistent, "persistent goroutine(s) found")
	println(r.temporary, "temporary goroutine(s) ignored")

	for _, g := range r.output {
		fmt.Println(r.traceGroups[g.traceFiltered], "goroutine(s) with similar back trace path")
		fmt.Println(g.id, g.status)
		fmt.Println(g.trace)
	}
}

func (r *res) fetch(n int, url string, delay time.Duration) {
	for i := 0; i < n; i++ {
		println("Collecting report ...")

		resp, err := http.Get(url)
		if err != nil {
			println("Failed to get report:", err)
			os.Exit(1)
		}

		parseGoroutines(resp.Body, r.result)

		err = resp.Body.Close()
		if err != nil {
			println("Failed to close response body:", err)
			os.Exit(1)
		}

		if i < n-1 {
			println("Sleeping", delay.String(), "...")
			time.Sleep(delay)
		}
	}
}

func (r *res) load(fn string) {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err.Error())
	}

	parseGoroutines(f, r.result)

	err = f.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (r *res) count() {
	for _, g := range r.result {
		if g.count > r.maxCount {
			r.maxCount = g.count
		}
	}

	r.traceGroups = make(map[string]int)
	r.output = make([]goroutine, 0, len(r.result))

	for _, g := range r.result {
		if g.count == r.maxCount {
			r.countPersistent(g)
		} else {
			r.temporary++
		}
	}

	sort.Slice(r.output, func(i, j int) bool {
		return r.output[i].traceFiltered < r.output[j].traceFiltered
	})
}

func (r *res) countPersistent(g goroutine) {
	r.persistent++

	if r.noGroup {
		r.output = append(r.output, g)
	} else {
		if _, ok := r.traceGroups[g.traceFiltered]; !ok {
			r.output = append(r.output, g)
		}
	}
	r.traceGroups[g.traceFiltered]++
}

type res struct {
	result map[string]goroutine

	maxCount    int
	noGroup     bool
	traceGroups map[string]int
	output      []goroutine
	persistent  int
	temporary   int
}

type goroutine struct {
	id            string
	count         int
	status        string
	trace         string
	traceFiltered string
}

var zeroX = regexp.MustCompile(`0x[a-z\d]+`)

func parseGoroutines(reader io.Reader, result map[string]goroutine) {
	g := goroutine{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "goroutine"):
			pieces := strings.SplitN(line, " ", 3)
			g.count = 1
			g.id = pieces[1]
			g.status = pieces[2]
			g.trace = ""
		case len(line) == 0:
			if gf, ok := result[g.id]; ok {
				g.count += gf.count
			}

			g.traceFiltered = zeroX.ReplaceAllString(g.trace, `0x?`)
			result[g.id] = g
		default:
			g.trace += line + "\n"
		}
	}
}
