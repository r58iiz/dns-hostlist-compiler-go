package cli

import (
	"flag"
	"fmt"
)

func ParseArgs() (string, string) {
	input := flag.String("input", "list.txt", "path to input list of URLs/files")
	output := flag.String("output", "outfile.txt", "path to output combined rules file")
	flag.Parse()

	if *input == "" {
		fmt.Println("input cannot be empty")
		flag.Usage()
		return "list.txt", *output
	}

	return *input, *output
}
