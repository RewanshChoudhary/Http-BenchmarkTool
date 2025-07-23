/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

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
		if (url==""){
			fmt.Println("A url is expected")
		}
		if (method==""){
			fmt.Println("The method type is expected")


		}
		runBenchmarkTool(url,requests,workers,method)


	},
}

func runBenchmarkTool(url string ,requests int ,workers int ,method string){
	
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.PersistentFlags().StringVar(&url,"url","","The url needed for the operation")
	benchmarkCmd.PersistentFlags().IntVar(&requests,"requests",1,"The number of requests for the endpoint you want to make Default: runs one 1 request")
    benchmarkCmd.PersistentFlags().IntVar(&workers,"concurrency",1,"The number of concurrrent workers you want to assign Default: Assigns 1 worker only")

	benchmarkCmd.PersistentFlags().StringVar(&method,"method","","The type of HTTP request is it For ex : Get , Post etc")




	
}
