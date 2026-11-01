package main

import (
	"os"

	"github.com/yemresalcan/mr-ducksafe/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
