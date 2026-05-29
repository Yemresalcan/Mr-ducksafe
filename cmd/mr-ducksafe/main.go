package main

import (
	"fmt"
	"os"

	"github.com/yemresalcan/mr-ducksafe/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
