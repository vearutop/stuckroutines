package stuckroutines

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof" // Introspecting goroutines.
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

func NewProcessor() *Processor {
	return &Processor{
		Result: make(map[string]goroutine),
	}
}

func Run(f Flags) {
	if f.URL == "" && flag.NArg() == 0 {
		flag.Usage()

		return
	}

	p := NewProcessor()
	p.NoGroup = f.NoGroup
	p.KeepTemporary = f.KeepTemporary

	if f.URL != "" {
		p.Fetch(f.Iterations, f.URL, f.Delay)
	}

	for _, fn := range flag.Args() {
		p.Load(fn)
	}

	p.Report(f)
}

func (p *Processor) Report(f Flags) {
	p.Count()
	p.PrepareOutput(f.SortTrace)
	p.PrintResult(f.MinCount)
}

func (p *Processor) PrintResult(minCount int) {
	fmt.Println(p.Persistent, "persistent goroutine(s) found")
	fmt.Println(p.Temporary, "temporary goroutine(s) ignored")

	for _, g := range p.Output {
		if minCount > 0 {
			if g.Count < minCount {
				continue
			}
		}

		fmt.Println(g.Count, "goroutine(s) with similar back trace path")
		fmt.Println(g.ID, g.Status)
		fmt.Println(g.Trace)
	}
}

func (p *Processor) PrepareOutput(sortTrace bool) {
	for i, g := range p.Output {
		g.Count = p.TraceGroups[g.TraceFiltered]

		p.Output[i] = g
	}

	if sortTrace {
		sort.Slice(p.Output, func(i, j int) bool {
			return p.Output[i].TraceFiltered < p.Output[j].TraceFiltered
		})
	} else {
		sort.Slice(p.Output, func(i, j int) bool {
			return p.Output[i].Count > p.Output[j].Count
		})
	}
}

func (p *Processor) FetchInternal() {
	if !p.internalStarted {
		p.internalStarted = true

		go func() {
			log.Println(http.ListenAndServe("localhost:45678", nil))
		}()
	}

	resp, err := http.Get("http://localhost:45678/debug/pprof/goroutine?debug=2")
	if err != nil {
		fmt.Println("Failed to get report:", err.Error())
		os.Exit(1)
	}

	parseGoroutines(resp.Body, p.Result)

	err = resp.Body.Close()
	if err != nil {
		fmt.Println("Failed to close response body:", err.Error())
		os.Exit(1)
	}
}

func (p *Processor) Fetch(n int, url string, delay time.Duration) {
	for i := 0; i < n; i++ {
		fmt.Println("Collecting report ...")

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Failed to get report:", err.Error())
			os.Exit(1)
		}

		parseGoroutines(resp.Body, p.Result)

		err = resp.Body.Close()
		if err != nil {
			fmt.Println("Failed to close response body:", err.Error())
			os.Exit(1)
		}

		if i < n-1 {
			fmt.Println("Sleeping", delay.String(), "...")
			time.Sleep(delay)
		}
	}
}

func (p *Processor) Load(fn string) {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err.Error())
	}

	parseGoroutines(f, p.Result)

	err = f.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (p *Processor) Count() {
	for _, g := range p.Result {
		if g.dumps > p.maxCount {
			p.maxCount = g.dumps
		}
	}

	p.TraceGroups = make(map[string]int)
	p.Output = make([]goroutine, 0, len(p.Result))

	for _, g := range p.Result {
		if g.dumps == p.maxCount || p.KeepTemporary {
			p.countPersistent(g)
		} else {
			p.Temporary++
		}
	}
}

func (p *Processor) countPersistent(g goroutine) {
	p.Persistent++

	if p.NoGroup {
		p.Output = append(p.Output, g)
	} else {
		if _, ok := p.TraceGroups[g.TraceFiltered]; !ok {
			p.Output = append(p.Output, g)
		}
	}
	p.TraceGroups[g.TraceFiltered]++
}

type Processor struct {
	Result map[string]goroutine

	maxCount      int
	NoGroup       bool
	TraceGroups   map[string]int
	Output        []goroutine
	Persistent    int
	Temporary     int
	KeepTemporary bool

	internalStarted bool
}

type goroutine struct {
	ID            string
	dumps         int
	Count         int
	Status        string
	Trace         string
	TraceFiltered string
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
			g.dumps = 1
			g.ID = pieces[1]
			g.Status = pieces[2]
			g.Trace = ""
		case len(line) == 0:
			if gf, ok := result[g.ID]; ok {
				g.dumps += gf.dumps
			}

			g.TraceFiltered = zeroX.ReplaceAllString(g.Trace, `0x?`)
			result[g.ID] = g
		default:
			g.Trace += line + "\n"
		}
	}
}
