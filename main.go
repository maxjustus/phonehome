package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/snadrus/metaphone3"
)

func main() {
	fmt.Println("starting")
	scanner := bufio.NewScanner(os.Stdin)

	encoder := metaphone3.New()
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
