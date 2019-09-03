package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	url := flag.String("url", "", "Full URL to /debug/pprof/goroutine?debug=2")
	n := flag.Int("iterations", 2, "How many reports to collect to find persisting routines")
	delay := flag.Duration("delay", 5*time.Second, "Delay between report collections")

	flag.Parse()

	result := make(map[string]goroutine)
	if *url == "" {
		flag.Usage()
		return
	}

	for i := 0; i < *n; i++ {
		println("Collecting report...")
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
			println("Sleeping...")
			time.Sleep(*delay)
		}
	}
	maxCount := 0
	for _, g := range result {
		if g.count > maxCount {
			maxCount = g.count
		}
	}
	removed := 0
	for _, g := range result {
		if g.count == maxCount {
			fmt.Println(g.id, g.status, "\n", g.trace)
		} else {
			removed++
		}
	}
	println(removed, "temporary goroutine(s) removed from report")
}

type goroutine struct {
	id     string
	count  int
	status string
	trace  string
}

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
			result[g.id] = g
		} else {
			g.trace += line + "\n"
		}
	}
}
