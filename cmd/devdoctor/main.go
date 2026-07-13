package main

import (
	"fmt"
	"os"

	"github.com/mahmudddddd/DevDoctor/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "devdoctor:", err)
		os.Exit(1)
	}
}
