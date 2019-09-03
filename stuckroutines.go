package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {
	url := flag.String("url", "", "Full URL to /debug/pprof/goroutine?debug=2")
	n := flag.Int("iterations", 2, "How many reports to collect to find persisting routines")
	delay := flag.Duration("delay", 5*time.Second, "Delay between report collections")
	noGroup := flag.Bool("no-group", false, "Do not group goroutines by stack trace")

	flag.Parse()

	result := make(map[string]goroutine)
	if *url == "" {
		flag.Usage()
		return
	}

	for i := 0; i < *n; i++ {
		println("Collecting report ...")
		resp, err := http.DefaultClient.Get(*url)
		if err != nil {
			log.Fatal(err.Error())
		}

		parseGoroutines(resp.Body, result)
		err = resp.Body.Close()
		if err != nil {
			log.Fatal(err.Error())
		}

		if i < *n-1 {
			println("Sleeping", delay.String(), "...")
			time.Sleep(*delay)
		}
	}
	maxCount := 0
	for _, g := range result {
		if g.count > maxCount {
			maxCount = g.count
		}
	}
	temporary := 0
	persistent := 0

	traceGroups := make(map[string]int)
	output := make([]goroutine, 0, len(result))
	for _, g := range result {
		if g.count == maxCount {
			persistent++

			if *noGroup {
				output = append(output, g)
			} else {
				if _, ok := traceGroups[g.traceFiltered]; !ok {
					output = append(output, g)
				}
			}
			traceGroups[g.traceFiltered]++
		} else {
			temporary++
		}
	}

	println(persistent, "persistent goroutine(s) found")
	println(temporary, "temporary goroutine(s) ignored")

	sort.Slice(output, func(i, j int) bool {
		return output[i].traceFiltered < output[j].traceFiltered
	})

	for _, g := range output {
		fmt.Println(traceGroups[g.traceFiltered], "goroutine(s) with similar back trace path")
		fmt.Println(g.id, g.status)
		fmt.Println(g.trace)
	}

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

		if strings.HasPrefix(line, "goroutine") {
			pieces := strings.SplitN(line, " ", 3)
			g.count = 1
			g.id = pieces[1]
			g.status = pieces[2]
			g.trace = ""
		} else if len(line) == 0 {
			if gf, ok := result[g.id]; ok {
				g.count += gf.count
			}
			g.traceFiltered = zeroX.ReplaceAllString(g.trace, `0x?`)
			result[g.id] = g
		} else {
			g.trace += line + "\n"
		}
	}
}
