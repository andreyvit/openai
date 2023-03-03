package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("Usage: convert-encoder-json encoder.json encoder.bin")
	}

	raw := must(os.ReadFile(flag.Arg(0)))

	var tokenEncodings map[string]int
	ensure(json.Unmarshal(raw, &tokenEncodings))

	var max int
	for _, v := range tokenEncodings {
		if v > max {
			max = v
		}
	}

	tokens := make([]string, max+1)
	for k, v := range tokenEncodings {
		tokens[v] = k
	}

	var buf bytes.Buffer
	for _, token := range tokens {
		buf.WriteString(token)
		buf.WriteByte(0)
	}

	ensure(os.WriteFile(flag.Arg(1), buf.Bytes(), 0644))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}
