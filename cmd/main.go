package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/pavelmemory/jobtome/internal"
)

func main() {
	var showVersionLong = flag.Bool("version", false, "")
	var showVersionShort = flag.Bool("v", false, "")
	flag.Parse()

	if *showVersionLong || *showVersionShort {
		_ = internal.WriteVersion(json.NewEncoder(os.Stdout))
		os.Exit(0)
	}

	if err := run(os.Args); err != nil {
		os.Exit(1)
	}
}
