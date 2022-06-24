# stuckroutines

A tool to retrieve long-running goroutines from pprof full goroutine stack dump that can help identifying stuck goroutines.

## Install

```
go install github.com/vearutop/stuckroutines@latest
$(go env GOPATH)/bin/stuckroutines --help
```

Or (with `go1.15` or older)

```
go get -u github.com/vearutop/stuckroutines
$(go env GOPATH)/bin/stuckroutines --help
```

Or download binary from [releases](https://github.com/vearutop/stuckroutines/releases).
```
wget https://github.com/vearutop/stuckroutines/releases/download/v1.1.0/linux_amd64.tar.gz && tar xf linux_amd64.tar.gz && rm linux_amd64.tar.gz
./stuckroutines -version
```

## Usage 

Make sure your app is instrumented with [net/pprof](https://golang.org/pkg/net/http/pprof/).

```
Stuckroutines requires either a URL or a list of files obtained from /pprof/goroutine?debug=2
Usage: stuckroutines [options] [...report files]
Usage of stuckroutines:
  -delay duration
        Delay between report collections (default 5s)
  -iterations int
        How many reports to collect to find persisting routines (default 2)
  -min-count int
        Filter traces with few goroutines (default 10)
  -no-group
        Do not group goroutines by stack trace
  -sort-trace
        Sort by trace instead of count ouf goroutines
  -url string
        Full URL to /debug/pprof/goroutine?debug=2
  -version
        Print version
```

Assuming your app is exposing `pprof` handlers at `http://my-service.acme.io/debug/pprof`:
```
stuckroutines -url http://my-service.acme.io/debug/pprof/goroutine?debug=2 > report.txt
```

```
Collecting report ...
Sleeping 5s ...
Collecting report ...
27 persistent goroutine(s) found
312 temporary goroutine(s) ignored
```

The report will only contain goroutines that have persisted between delayed iterations.

Alternatively if you already have downloaded goroutine dumps you can provide the files:
```
stuckroutines dump1.txt dump2.txt dump3.txt > report.txt
```

List of goroutines is ordered by count of goroutines to show the likely culprit,
it can be also ordered by back trace path (with memory references removed) to be diff-friendly.

<details>
  <summary>Sample report</summary>

```
10 goroutine(s) with similar back trace path
73 [select, 30 minutes]:
database/sql.(*DB).connectionOpener(0xc0001f2900, 0xf61260, 0xc000176cc0)
	/usr/local/go/src/database/sql/sql.go:1000 +0xe8
created by database/sql.OpenDB
	/usr/local/go/src/database/sql/sql.go:670 +0x15e

11 goroutine(s) with similar back trace path
86 [select, 12 minutes]:
database/sql.(*DB).connectionResetter(0xc0001f2d80, 0xf61260, 0xc000176e40)
	/usr/local/go/src/database/sql/sql.go:1013 +0xfb
created by database/sql.OpenDB
	/usr/local/go/src/database/sql/sql.go:671 +0x194

1 goroutine(s) with similar back trace path
1 [select, 30 minutes]:
github.com/acme/my-service/cmd.glob..func5.1(0x0, 0x0, 0xc00003c06a, 0x7, 0xc00003c1ab, 0x8, 0xc0005d2ab0, 0xd84c2b, 0x6, 0x0, ...)
	/go/src/github.com/acme/my-service/cmd/web.go:189 +0x295
reflect.Value.call(0xc4b420, 0xe496e0, 0x13, 0xd82085, 0x4, 0xc0003cbaa0, 0x1, 0x1, 0x1, 0xc00017f110, ...)
	/usr/local/go/src/reflect/value.go:447 +0x461
reflect.Value.Call(0xc4b420, 0xe496e0, 0x13, 0xc0003cbaa0, 0x1, 0x1, 0xc00014a880, 0xc0003cbaa0, 0x1)
	/usr/local/go/src/reflect/value.go:308 +0xa4
github.com/acme/my-service/vendor/go.uber.org/dig.(*Container).Invoke(0xc00014a880, 0xc4b420, 0xe496e0, 0x0, 0x0, 0x0, 0x4, 0x820d20)
	/go/src/github.com/acme/my-service/vendor/go.uber.org/dig/dig.go:518 +0x3d8
github.com/acme/my-service/cmd.glob..func5(0x15d6e40, 0x15fede0, 0x0, 0x0)
	/go/src/github.com/acme/my-service/cmd/web.go:168 +0x5a
github.com/acme/my-service/vendor/github.com/spf13/cobra.(*Command).execute(0x15d6e40, 0x15fede0, 0x0, 0x0, 0x15d6e40, 0x15fede0)
	/go/src/github.com/acme/my-service/vendor/github.com/spf13/cobra/command.go:766 +0x2ae
github.com/acme/my-service/vendor/github.com/spf13/cobra.(*Command).ExecuteC(0x15d6980, 0x1, 0x0, 0x0)
	/go/src/github.com/acme/my-service/vendor/github.com/spf13/cobra/command.go:850 +0x2fc
github.com/acme/my-service/vendor/github.com/spf13/cobra.(*Command).Execute(...)
	/go/src/github.com/acme/my-service/vendor/github.com/spf13/cobra/command.go:800
github.com/acme/my-service/cmd.Execute()
	/go/src/github.com/acme/my-service/cmd/root.go:12 +0x32
main.main()
	/go/src/github.com/acme/my-service/main.go:6 +0x20

1 goroutine(s) with similar back trace path
99 [chan receive, 30 minutes]:
github.com/acme/my-service/pkg/config.EnableGracefulShutdown.func1.1(0x0, 0x0, 0xf61260, 0xc000176700, 0xc00015f550, 0xc000149ad0, 0xf73740, 0xc00014e0a0, 0xc000153800, 0xc000153800, ...)
	/go/src/github.com/acme/my-service/pkg/config/router.go:168 +0x173
created by github.com/acme/my-service/pkg/config.EnableGracefulShutdown.func1
	/go/src/github.com/acme/my-service/pkg/config/router.go:163 +0x127

1 goroutine(s) with similar back trace path
56 [select, 30 minutes]:
github.com/acme/my-service/vendor/github.com/acme/root-kit/pkg/cache.(*Memory).cleaner(0xc0001b6140)
	/go/src/github.com/acme/my-service/vendor/github.com/acme/root-kit/pkg/cache/memory.go:208 +0xf5
created by github.com/acme/my-service/vendor/github.com/acme/root-kit/pkg/cache.NewMemory
	/go/src/github.com/acme/my-service/vendor/github.com/acme/root-kit/pkg/cache/memory.go:99 +0x210

1 goroutine(s) with similar back trace path
116 [select]:
github.com/acme/my-service/vendor/github.com/streadway/amqp.(*Connection).heartbeater(0xc0000d4b40, 0x2540be400, 0xc00018e4e0)
	/go/src/github.com/acme/my-service/vendor/github.com/streadway/amqp/connection.go:551 +0x187
created by github.com/acme/my-service/vendor/github.com/streadway/amqp.(*Connection).openTune
	/go/src/github.com/acme/my-service/vendor/github.com/streadway/amqp/connection.go:782 +0x482

1 goroutine(s) with similar back trace path
35 [select]:
github.com/acme/my-service/vendor/go.opencensus.io/stats/view.(*worker).start(0xc0001b41e0)
	/go/src/github.com/acme/my-service/vendor/go.opencensus.io/stats/view/worker.go:154 +0x100
created by github.com/acme/my-service/vendor/go.opencensus.io/stats/view.init.0
	/go/src/github.com/acme/my-service/vendor/go.opencensus.io/stats/view/worker.go:32 +0x57

1 goroutine(s) with similar back trace path
36 [syscall, 30 minutes]:
os/signal.signal_recv(0x0)
	/usr/local/go/src/runtime/sigqueue.go:139 +0x9c
os/signal.loop()
	/usr/local/go/src/os/signal/signal_unix.go:23 +0x22
created by os/signal.init.0
	/usr/local/go/src/os/signal/signal_unix.go:29 +0x41

```
</details>
