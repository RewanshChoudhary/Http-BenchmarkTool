# Http-BenchmarkTool

A lightweight CLI benchmarking tool written in Go for testing HTTP endpoints with concurrency support, custom headers.

**Performance Summary**

  After the run, you'll get:
  - Total, min, max, median, and average latency
  - Success vs. failure count
  - Per-request timing breakdown
  - HTTP status code distribution

 **JSON Summary Export**  
  Set `--prettyjson=false` to avoiding export a structured `summary.json` file with full benchmark metrics.
  Default is false to ensure only clean result is provided

  ->Automatically detects `@file` syntax for external JSON files.
  
**Header Validation**
  Ensures proper formatting of custom headers.

**Thread-Safe HTTP Response Stats**
  Uses a `sync.Mutex`-protected map to prevent race conditions when counting HTTP status codes.

**Error Handling**  
  Gracefully handles:
  - Invalid headers
  - File read errors
  - Failed HTTP requests
  - Timeout cases


##  Installation  

```bash
go install github.com/RewanshChoudhary/Http-BenchmarkTool@latest
```


Usage guide:
```bash
Http-BenchmarkTool benchmark \
  --url https://example.com \
  --requests 100 \
  --workers 10 \
  --method GET \
  --timeout 5 \
  --headers "Authorization: Bearer token" \
  --headers "User-Agent: CustomAgent"
```
 Short Description of flags:-

| Flag         | Description                       |
| ------------ | --------------------------------- |
| `--url`      | Target URL                        |
| `--requests` | Total number of requests to send  |
| `--workers`  | Number of concurrent workers      |
| `--method`   | HTTP method (`GET`, `POST`, etc.) |
| `--timeout`  | Timeout in seconds                |
| `--headers`  | Custom headers (can be repeated)  |



Structure:

Http-BenchmarkTool/
├── cmd/
│ └── benchmark.go
├── go.mod
├── go.sum
├── main.go
├── README.md
└── LICENSE



