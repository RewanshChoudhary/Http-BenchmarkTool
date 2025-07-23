/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var url string
var requests int

var workers int
var method string

// benchmarkCmd represents the benchmark command
var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Provides core flags to test",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

func worker(url string, jobs chan int, workers int, method string, results chan<- response, wg *sync.WaitGroup) {
	client := http.Client{}

	defer wg.Done()

	for range jobs {
		start := time.Now()

		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			results <- response{0, false}

			continue

		}
		resp, err := client.Do(req)
		elapsed := time.Since(start)

		if err != nil {
			results <- response{elapsed, false}

		}
		success := resp.StatusCode >= 200 && resp.StatusCode < 300

		results <- response{elapsed, success}

	}

}

func runBenchmarkTool(url string, requests int, workers int, method string) {
	var wg sync.WaitGroup

	jobs := make(chan int, requests)

	results := make(chan response, requests)

	for i := 0; i < workers; i++ {

		wg.Add(1)

		go worker(url, jobs, workers, method, results, &wg)

	}

	for i := 0; i < requests; i++ {
		jobs <- i
	}

	close(jobs)

	wg.Wait()

	close(results)

	var min ,max time.Duration
	var totalDelay time.Duration
	var success ,failed int


	for res:=range results{
		if (res.success){
			success++
		}else {
			failed++

		}
		totalDelay+=res.delay

		if(res.delay<min){
			min=res.delay

		}
		if (res.delay>max){
			max=res.delay
		}

	}
	avgTimeTaken:=totalDelay/time.Duration(requests)

	fmt.Println("\n--- Benchmark Summary ---")
	fmt.Printf("Total Requests: %d\n", requests)
	fmt.Printf("Successful: Attempts      %d\n", success)
	fmt.Printf("Failed: Attempts        %d\n", failed)
	fmt.Printf("Avg Latency:    %v\n", avgTimeTaken)
	fmt.Printf("Min Latency:    %v\n", min)
	fmt.Printf("Max Latency:    %v\n", max)
	
	

}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.PersistentFlags().StringVar(&url, "url", "", "The url needed for the operation")
	benchmarkCmd.PersistentFlags().IntVar(&requests, "requests", 1, "The number of requests for the endpoint you want to make Default: runs one 1 request")
	benchmarkCmd.PersistentFlags().IntVar(&workers, "concurrency", 1, "The number of concurrrent workers you want to assign Default: Assigns 1 worker only")

	benchmarkCmd.PersistentFlags().StringVar(&method, "method", "", "The type of HTTP request is it For ex : Get , Post etc")

}
