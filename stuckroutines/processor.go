// Package stuckroutines analyzes goroutine dumps.
package stuckroutines

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof" // Introspecting goroutines.
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// NewProcessor creates Processor.
func NewProcessor() *Processor {
	return &Processor{
		Result: make(map[string]goroutine),
		Writer: os.Stdout,
	}
}

// Run invokes processing.
func Run(f Flags) {
	if f.URL == "" && flag.NArg() == 0 {
		flag.Usage()

		return
	}

	p := NewProcessor()
	p.NoGroup = f.NoGroup
	p.KeepTemporary = f.KeepTemporary
	p.f = f

	if f.URL != "" {
		p.Fetch(f.Iterations, f.URL, f.Delay)
	}

	for _, fn := range flag.Args() {
		p.Load(fn)
	}

	p.Report(f)
}

// Report processes and prints currently collected dumps.
func (p *Processor) Report(f Flags) {
	p.Count()
	p.PrepareOutput(f.SortTrace)
	p.PrintResult(f.MinCount)
}

// PrintResult prints grouped traces.
func (p *Processor) PrintResult(minCount int) {
	_, _ = fmt.Fprintln(p.Writer, p.Persistent, "persistent goroutine(s) found")
	_, _ = fmt.Fprintln(p.Writer, p.Temporary, "temporary goroutine(s) ignored")

	for _, g := range p.Output {
		if minCount > 0 {
			if g.Count < minCount {
				continue
			}
		}

		_, _ = fmt.Fprintln(p.Writer, g.Count, "goroutine(s) with similar back trace path")
		_, _ = fmt.Fprintln(p.Writer, g.ID, g.Status)

		trc := g.Trace
		if p.f.ShowFiltered {
			trc = g.TraceFiltered
		}

		if p.f.TruncateTrace > 0 {
			tr := strings.Split(trc, "\n")
			if len(tr) > p.f.TruncateTrace {
				tr = tr[:p.f.TruncateTrace]
				_, _ = fmt.Fprintln(p.Writer, strings.Join(tr, "\n")+"\n")
			} else {
				_, _ = fmt.Fprintln(p.Writer, trc)
			}
		} else {
			_, _ = fmt.Fprintln(p.Writer, trc)
		}
	}
}

// PrepareOutput filters and orders traces.
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

// Internal dumps goroutines of current process.
func (p *Processor) Internal() {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/debug/pprof/goroutine?debug=2", nil) //nolint:errcheck
	rw := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		panic(fmt.Sprintf("unexpected response: %d %s", rw.Code, rw.Body.String()))
	}

	p.parseGoroutines(rw.Body, p.Result)
}

// Fetch downloads multiple goroutine dumps by URL.
func (p *Processor) Fetch(n int, url string, delay time.Duration) {
	for i := 0; i < n; i++ {
		_, _ = fmt.Fprintln(p.Writer, "Collecting report ...")

		resp, err := http.Get(url)
		if err != nil {
			_, _ = fmt.Fprintln(p.Writer, "Failed to get report:", err.Error())

			os.Exit(1)
		}

		p.parseGoroutines(resp.Body, p.Result)

		err = resp.Body.Close()
		if err != nil {
			_, _ = fmt.Fprintln(p.Writer, "Failed to close response body:", err.Error())

			os.Exit(1)
		}

		if i < n-1 {
			_, _ = fmt.Fprintln(p.Writer, "Sleeping", delay.String(), "...")
			time.Sleep(delay)
		}
	}
}

// Load adds goroutines dump from file.
func (p *Processor) Load(fn string) {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err.Error())
	}

	p.parseGoroutines(f, p.Result)

	err = f.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Count counts persistent goroutines.
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

// Processor groups goroutine stack traces.
type Processor struct {
	Result map[string]goroutine

	maxCount      int
	NoGroup       bool
	TraceGroups   map[string]int
	Output        []goroutine
	Persistent    int
	Temporary     int
	KeepTemporary bool

	Writer io.Writer

	f Flags
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

func (p *Processor) parseGoroutines(reader io.Reader, result map[string]goroutine) {
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

			if p.f.TruncateTrace > 0 {
				tr := strings.Split(g.TraceFiltered, "\n")
				if len(tr) > p.f.TruncateTrace {
					tr = tr[:p.f.TruncateTrace]

					g.TraceFiltered = strings.Join(tr, "\n") + "\n"
				}
			}

			result[g.ID] = g
		default:
			g.Trace += line + "\n"
		}
	}
}
