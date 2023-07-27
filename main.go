package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/snadrus/metaphone3"
)

var cli struct {
	Vowels bool `help:"Makes metaphone3 encode non-initial vowels" default:"false"`
	Exact  bool `help:"Makes metaphone3 encode consonants as exactly as possible." default:"false"`
}

func main() {
	fmt.Fprintln(os.Stderr, "Listening on stdin")

	kong.Parse(&cli)

	scanner := bufio.NewScanner(os.Stdin)
	encoder := metaphone3.New()

	encoder.SetEncodeVowels(cli.Vowels)
	encoder.SetEncodeExact(cli.Exact)

	for scanner.Scan() {
		word := scanner.Text()

		if word == "" {
			// ClickHouse sends empty input sometimes, just ignore
			continue
		}

		primary, secondary := encoder.Encode(word)

		fmt.Printf("%s\t%s\n", primary, secondary)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
