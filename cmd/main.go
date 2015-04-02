package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/images"
)

const version = "0.0.1"

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

func realMain() int {
	var (
		flagProvider = flag.String("provider", "", "Cloud provider")
		flagVersion  = flag.Bool("version", false, "Show version and exit")
	)

	flag.Parse()
	if *flagVersion {
		return 0
	}

	conf := &images.Config{
		Provider: *flagProvider,
	}

	if err := images.Run(conf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
