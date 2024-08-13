package main

//usage: go run mtrparser.go <hostname or ip>

import (
	"fmt"
	"log"

	flag "github.com/spf13/pflag"
)

var flagInterval float64
var flagCount int
var flagTarget string
var flagIPVersion string

func init() {
	flag.Float64Var(&flagInterval, "interval", 1, "Interval (ms) between mtr packets")
	flag.IntVar(&flagCount, "count", 10, "Number of mtr packets to send")
	flag.StringVar(&flagTarget, "target", "", "Destination address")
	flag.StringVar(&flagIPVersion, "ip", "4", "IP version to use (4 or 6)")

	flag.Parse()
}

func main() {
	if flagTarget == "" {
		log.Fatal("Need mtr destination")
	}
	fmt.Println("Interval:", flagInterval, "Count:", flagCount)

	result, err := ExecuteMTR(flagTarget, flagIPVersion, flagCount, flagInterval)
	if err != nil {
		log.Fatal(err)
	}

	for _, hop := range result.Hops {
		fmt.Print(hop.Timings)
		hop.Summarize(flagCount)
	}

	// fmt.Println("mtr --report like output")
	fmt.Println(result.String())
}
