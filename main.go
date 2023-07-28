package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/snadrus/metaphone3"
)

var cli struct {
	Vowels bool `help:"Makes metaphone3 encode non-initial vowels" default:"false"`
	Exact  bool `help:"Makes metaphone3 encode consonants as exactly as possible." default:"false"`
}

type InputEvent struct {
	ReadOffset uint64
	Text       string
}

type OutputEvent struct {
	WriteOffset    uint64
	SerializedText string
}

func processInputs(inputs <-chan InputEvent, outputs chan<- OutputEvent) {
	poolSize := runtime.NumCPU()

	wg := sync.WaitGroup{}

	for i := 0; i < poolSize; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			encoder := metaphone3.New()

			encoder.SetEncodeVowels(cli.Vowels)
			encoder.SetEncodeExact(cli.Exact)

			for input := range inputs {
				primary, secondary := encoder.Encode(input.Text)

				text := fmt.Sprintf("['%s','%s']\n", primary, secondary)

				outputs <- OutputEvent{WriteOffset: input.ReadOffset, SerializedText: text}
			}
		}()
	}

	wg.Wait()
}

func main() {
	fmt.Fprintln(os.Stderr, "Listening on stdin")

	kong.Parse(&cli)

	scanner := bufio.NewScanner(os.Stdin)

	input := make(chan InputEvent, 1000)
	output := make(chan OutputEvent, 1000)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		readOffsetCounter := uint64(0)

		for scanner.Scan() {
			word := scanner.Text()

			if word == "" {
				// ClickHouse sends empty input sometimes, just ignore
				continue
			}

			input <- InputEvent{ReadOffset: readOffsetCounter, Text: word}
			readOffsetCounter++
		}

		wg.Done()

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		processInputs(input, output)

		close(output)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		writeOffsetCounter := uint64(0)
		writeMap := make(map[uint64]string)

		for outputEvent := range output {
			writeMap[outputEvent.WriteOffset] = outputEvent.SerializedText

			for {
				if text, ok := writeMap[writeOffsetCounter]; ok {
					fmt.Print(text)
					delete(writeMap, writeOffsetCounter)
					writeOffsetCounter++
				} else {
					break
				}
			}
		}
	}()

	wg.Wait()

	os.Exit(0)
}
