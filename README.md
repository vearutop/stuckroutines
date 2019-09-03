# stuckroutines

A tool to retrieve long-running goroutines from pprof full goroutine stack dump that can help identifying stuck goroutines.

## Installation

```
go get github.com/vearutop/stuckroutines
```

## Usage 

Make sure your app is instrumented with [net/pprof](https://golang.org/pkg/net/http/pprof/).

```
Usage of stuckroutines:
  -delay duration
        Delay between report collections (default 5s)
  -iterations int
        How many reports to collect to find persisting routines (default 2)
  -url string
        Full URL to /debug/pprof/goroutine?debug=2
```

Assuming your app is exposing `pprof` handlers at `http://127.0.0.1:8000/debug/pprof`:
```
stuckroutines -url http://127.0.0.1:8000/debug/pprof/goroutine?debug=2 > report.txt
```

```
Collecting report...
Sleeping...
Collecting report...
```

The report will only contain goroutines that have persisted between delayed iterations.

<details>
  <summary>Sample `report.txt`</summary>

```
1 [chan receive, 2 minutes]: 
 github.com/acme/root-kit/pkg/http/server.(*Instance).Start(0xc0000b00c0, 0xc000561db8, 0x17f3640)
	/Users/john.doe/dev/root-kit/pkg/http/server/server.go:69 +0x201
main.main()
	/Users/john.doe/dev/root-kit/examples/basic-service/cmd/basic-service/main.go:38 +0x97e

20 [select]: 
 go.opencensus.io/stats/view.(*worker).start(0xc000304230)
	/Users/john.doe/go/pkg/mod/go.opencensus.io@v0.22.0/stats/view/worker.go:154 +0x100
created by go.opencensus.io/stats/view.init.0
	/Users/john.doe/go/pkg/mod/go.opencensus.io@v0.22.0/stats/view/worker.go:32 +0x57

51 [syscall, 2 minutes]: 
 os/signal.signal_recv(0x0)
	/usr/local/Cellar/go/1.12.6/libexec/src/runtime/sigqueue.go:139 +0x9f
os/signal.loop()
	/usr/local/Cellar/go/1.12.6/libexec/src/os/signal/signal_unix.go:23 +0x22
created by os/signal.init.0
	/usr/local/Cellar/go/1.12.6/libexec/src/os/signal/signal_unix.go:29 +0x41

56 [IO wait]: 
 internal/poll.runtime_pollWait(0x1f24ea8, 0x72, 0x0)
	/usr/local/Cellar/go/1.12.6/libexec/src/runtime/netpoll.go:182 +0x56
internal/poll.(*pollDesc).wait(0xc00029a898, 0x72, 0x0, 0x0, 0x1693ae5)
	/usr/local/Cellar/go/1.12.6/libexec/src/internal/poll/fd_poll_runtime.go:87 +0x9b
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/Cellar/go/1.12.6/libexec/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Accept(0xc00029a880, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0)
	/usr/local/Cellar/go/1.12.6/libexec/src/internal/poll/fd_unix.go:384 +0x1ba
net.(*netFD).accept(0xc00029a880, 0x52edd4b4, 0x6c152edd4b4, 0x100000001)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/fd_unix.go:238 +0x42
net.(*TCPListener).accept(0xc000298008, 0x5d6e600e, 0xc0004cae60, 0x10b5ec6)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/tcpsock_posix.go:139 +0x32
net.(*TCPListener).Accept(0xc000298008, 0xc0004caeb0, 0x18, 0xc00008c780, 0x130d974)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/tcpsock.go:260 +0x48
net/http.(*Server).Serve(0xc0002c0000, 0x17fd980, 0xc000298008, 0x0, 0x0)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:2859 +0x22d
github.com/acme/root-kit/pkg/http/server.(*Instance).Start.func1(0xc0002c0000, 0x17fd980, 0xc000298008, 0xc0000b00c0)
	/Users/john.doe/dev/root-kit/pkg/http/server/server.go:53 +0x43
created by github.com/acme/root-kit/pkg/http/server.(*Instance).Start
	/Users/john.doe/dev/root-kit/pkg/http/server/server.go:52 +0x1b7

55 [chan receive, 2 minutes]: 
 github.com/acme/root-kit/pkg/http/server.(*Instance).handleServerShutdown.func1(0xc000030300, 0xc0000b00c0, 0xc0002c0000)
	/Users/john.doe/dev/root-kit/pkg/http/server/server.go:85 +0x38
created by github.com/acme/root-kit/pkg/http/server.(*Instance).handleServerShutdown
	/Users/john.doe/dev/root-kit/pkg/http/server/server.go:84 +0xdc

62 [running]: 
 runtime/pprof.writeGoroutineStacks(0x17f3020, 0xc0004220e0, 0x0, 0x0)
	/usr/local/Cellar/go/1.12.6/libexec/src/runtime/pprof/pprof.go:679 +0x9d
runtime/pprof.writeGoroutine(0x17f3020, 0xc0004220e0, 0x2, 0x100d7e9, 0xc0002b0160)
	/usr/local/Cellar/go/1.12.6/libexec/src/runtime/pprof/pprof.go:668 +0x44
runtime/pprof.(*Profile).WriteTo(0x1c54aa0, 0x17f3020, 0xc0004220e0, 0x2, 0xc0004220e0, 0xc0000befc0)
	/usr/local/Cellar/go/1.12.6/libexec/src/runtime/pprof/pprof.go:329 +0x390
net/http/pprof.handler.ServeHTTP(0xc00048c221, 0x9, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/pprof/pprof.go:245 +0x356
net/http/pprof.Index(0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/pprof/pprof.go:268 +0x6f7
net/http.HandlerFunc.ServeHTTP(0x1748278, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
net/http.(*ServeMux).ServeHTTP(0xc000470840, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:2375 +0x1d6
github.com/go-chi/chi.(*Mux).Mount.func1(0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/mux.go:292 +0x127
net/http.HandlerFunc.ServeHTTP(0xc0001de540, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/go-chi/chi.(*Mux).routeHTTP(0xc0000db020, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/mux.go:425 +0x27f
net/http.HandlerFunc.ServeHTTP(0xc000020de0, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/go-chi/chi.(*Mux).ServeHTTP(0xc0000db020, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/mux.go:70 +0x451
github.com/go-chi/chi.(*Mux).Mount.func1(0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/mux.go:292 +0x127
net/http.HandlerFunc.ServeHTTP(0xc0001de660, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/go-chi/chi.(*Mux).routeHTTP(0xc0000da180, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/mux.go:425 +0x27f
net/http.HandlerFunc.ServeHTTP(0xc000020680, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/go-chi/chi/middleware.RealIP.func1(0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/middleware/realip.go:34 +0x99
net/http.HandlerFunc.ServeHTTP(0xc0001de120, 0x17fdc40, 0xc0004220e0, 0xc00045a400)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/go-chi/render.SetContentType.func1.1(0x17fdc40, 0xc0004220e0, 0xc00045a300)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/render@v1.0.1/content_type.go:52 +0x18b
net/http.HandlerFunc.ServeHTTP(0xc0001de140, 0x17fdc40, 0xc0004220e0, 0xc00045a300)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/acme/root-kit/pkg/http/router.(*PanicRecoverer).Middleware.func1(0x17fdc40, 0xc0004220e0, 0xc00045a300)
	/Users/john.doe/dev/root-kit/pkg/http/router/panic.go:30 +0xa8
net/http.HandlerFunc.ServeHTTP(0xc0001de160, 0x17fdc40, 0xc0004220e0, 0xc00045a300)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1995 +0x44
github.com/go-chi/chi.(*Mux).ServeHTTP(0xc0000da180, 0x17fdc40, 0xc0004220e0, 0xc00045a200)
	/Users/john.doe/go/pkg/mod/github.com/go-chi/chi@v4.0.2+incompatible/mux.go:82 +0x294
net/http.serverHandler.ServeHTTP(0xc0002c0000, 0x17fdc40, 0xc0004220e0, 0xc00045a200)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:2774 +0xa8
net/http.(*conn).serve(0xc0001e7220, 0x17ff780, 0xc0000ce100)
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:1878 +0x851
created by net/http.(*Server).Serve
	/usr/local/Cellar/go/1.12.6/libexec/src/net/http/server.go:2884 +0x2f4

```
</details>
