package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	timeout  int
	url      string
	requests int

	workers    int
	method     string
	headers    []string
	body       string
	prettyJson bool
)

type PerfSummary struct {
	Scale                     string
	TotalRequests             int
	ConcurrentWorkersAssigned int
	HeadersProvided           []string
	BodyProvided              string

	TotalDelay       time.Duration
	Median           time.Duration
	AverageTimeTaken time.Duration
	SuccessCount     int
	FailedCount      int
	MinLatency       time.Duration
	MaxLatency       time.Duration
	DelayPerRequest  []time.Duration
}
type SafeMap struct {
	mu sync.Mutex
	m  map[int]int
	// helps preventing deadlock between GoRountines
	// which trys to update and access at the same time
}

// benchmarkCmd represents the benchmark command
var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Provides core flags to test your http endpoints ",
	Long:  "Helps in examining and checking performance of endpoints ",
	Run: func(cmd *cobra.Command, args []string) {
		if url == "" {
			fmt.Println("A url is expected")
		}
		if method == "" {
			fmt.Println("The method type is expected")

		}
		runBenchmarkTool(url, requests, workers, method)

	},
}

type response struct {
	delay   time.Duration
	success bool
}

func printStatusCodeCount(resultMap map[int]int) {
	fmt.Println("Status Code :  The Count of it ")
	for key, value := range resultMap {

		fmt.Println(key, "-->", value)

	}
}
func addHeaders(req *http.Request, headers []string) {
	for _, value := range headers {
		separated := strings.SplitN(value, ":", 2)
		if len(separated) != 2 {
			fmt.Printf("Header's are wrong  ")
			continue

		}
		key := strings.TrimSpace(separated[0])
		value := strings.TrimSpace(separated[1])

		req.Header.Add(key, value)

	}
}

func startWorkers(workers int, jobs chan int, results chan<- response, url, method string, statusCounter *SafeMap, wg *sync.WaitGroup) {

	for i := 0; i < workers; i++ {

		wg.Add(1)

		go worker(timeout, statusCounter, url, jobs, method, results, wg)

	}

}

// Add application/json header for a json text
func understandBody(body string) io.Reader {
	if body == "" {
		return nil
	}

	data := strings.TrimLeft(body, " ")
	if data[0] == '@' {
		filePath := data[1:]
		info, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Encountered an error while reading the file ")
			return nil
		}
		return bytes.NewReader(info)
	}

	info := strings.TrimSpace(body)

	return strings.NewReader(info)

}

func worker(timeout int, countStatusCode *SafeMap, url string, jobs chan int, method string, results chan<- response, wg *sync.WaitGroup) {
	client := http.Client{}

	defer wg.Done()

	for range jobs {
		start := time.Now()

		req, err := http.NewRequest(method, url, understandBody(body))
		if err != nil {
			results <- response{0, false}

			continue

		}

		if body != "" {
			req.Header.Add("Content-type", "application/json")

		}

		addHeaders(req, headers)

		understandBody(body)

		resp, err := client.Do(req)
		elapsed := time.Since(start)
		defer resp.Body.Close()

		if err != nil {
			results <- response{elapsed, false}
			continue

		}
		if err == nil {
			countStatusCode.mu.Lock()
			countStatusCode.m[resp.StatusCode]++
			countStatusCode.mu.Unlock()
			defer resp.Body.Close()
		}

		if int(elapsed.Milliseconds()) >= timeout {
			results <- response{elapsed, false}
			continue

		}

		success := resp.StatusCode >= 200 && resp.StatusCode < 300

		results <- response{elapsed, success}

	}

}
func sortSlice(delayPerRequest []time.Duration) {
	sort.Slice(delayPerRequest, func(i, j int) bool {
		return delayPerRequest[i] < delayPerRequest[j]

	})
}
func properties(results chan response, success, failed *int, totalDelay, min, max *time.Duration) []time.Duration {
	var delayPerRequest []time.Duration
	for res := range results {
		delayPerRequest = append(delayPerRequest, (res.delay))

		if res.success {
			*success++
		} else {
			*failed++

		}
		*totalDelay += res.delay

		if res.delay < *min {
			*min = res.delay

		}
		if res.delay > *max {
			*max = res.delay
		}

	}
	return delayPerRequest
}

func exportJsonFile(jsonStruct *PerfSummary) {
	jsonData, err := json.MarshalIndent(jsonStruct, "", " ")
	if err != nil {
		log.Fatalf("Uncountered an error while identing the file %v", err)
	}
	er := os.WriteFile("summary.json", jsonData, 0644)
	if er != nil {
		log.Fatalf("Uncountered an error while making a file %v", er)

	}

	fmt.Println("Json file was written: ")

}

func runBenchmarkTool(url string, requests int, workers int, method string) {
	var wg sync.WaitGroup
	counter := &SafeMap{m: make(map[int]int)}

	delayPerRequest := []time.Duration{}

	jobs := make(chan int, requests)

	results := make(chan response, requests)
	startWorkers(workers, jobs, results, url, method, counter, &wg)

	for i := 0; i < requests; i++ {
		jobs <- i
	}

	close(jobs)

	wg.Wait()

	close(results)

	var min, max time.Duration
	var totalDelay time.Duration
	var success, failed int
	var median time.Duration

	delayPerRequest = properties(results, &success, &failed, &totalDelay, &min, &max)

	sortSlice(delayPerRequest)

	l := len(delayPerRequest)

	if l%2 == 0 {
		median = (delayPerRequest[l/2] + delayPerRequest[(l/2)-1]) / 2

	} else {
		median = (delayPerRequest[l/2])

	}
	avgTimeTaken := totalDelay / time.Duration(requests)
	if prettyJson {
		results := PerfSummary{"NanoSeconds",
			requests,
			workers,
			headers,
			body,

			totalDelay,
			median,
			avgTimeTaken,
			success,
			failed,
			min,
			max,
			delayPerRequest,
		}
		exportJsonFile(&results)
	}
	if prettyJson == false {
		fmt.Println("\n--- Benchmark Summary ---")
		fmt.Printf("Total Requests: %d\n", requests)
		fmt.Printf("Successful: Attempts      %d\n", success)
		fmt.Printf("Failed: Attempts        %d\n", failed)
		fmt.Printf("Avg Latency:    %v\n", avgTimeTaken)
		fmt.Printf("Min Latency:    %v\n", min)
		fmt.Printf("Max Latency:    %v\n", max)
		fmt.Println(delayPerRequest)
		fmt.Printf("The median the value of latency is : %v\n", median)

		fmt.Println("The count of each status code is as follows ")
		printStatusCodeCount(counter.m)
	}
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.PersistentFlags().StringVar(&url, "url", "", "The url needed for the operation")
	benchmarkCmd.PersistentFlags().IntVar(&requests, "requests", 1, "The number of requests for the endpoint you want to make Default: runs one 1 request")
	benchmarkCmd.PersistentFlags().IntVar(&workers, "concurrency", 1, "The number of concurrrent workers you want to assign Default: Assigns 1 worker only")
	benchmarkCmd.PersistentFlags().StringVar(&method, "method", "", "The type of HTTP request is it For ex : Get , Post etc")
	benchmarkCmd.PersistentFlags().IntVar(&timeout, "timeout", 10000, "The timeout set for each request int miliseconds(ms) ")
	benchmarkCmd.PersistentFlags().StringArrayVarP(&headers, "header", "H", []string{}, "Custom headers to include in the requests")
	benchmarkCmd.PersistentFlags().StringVar(&body, "body", "", "The body that is required for the endpoint a json string or @file (e.g. --body='{\"key\":\"value\"}' or --body=@data.json)")
	benchmarkCmd.PersistentFlags().BoolVar(&prettyJson, "prettyjson", true, "Provides a clean json format summary with each requests time delay ")
	

}
