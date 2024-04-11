package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/andreyvit/openai"
)

var (
	alwaysShowPaths = flag.Bool("f", false, "print file names even if only a single file is specified")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	showPaths := *alwaysShowPaths || flag.NArg() > 1

	var total int
	for _, fn := range flag.Args() {
		raw := string(must(os.ReadFile(fn)))
		count := openai.TokenCount(raw, openai.ModelChatGPT4)
		total += count
		if showPaths {
			fmt.Printf("%s: %d\n", filepath.Base(fn), count)
		} else {
			fmt.Println(count)
		}
	}
	if flag.NArg() > 1 {
		fmt.Printf("TOTAL: %d\n", total)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
