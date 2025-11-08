package main

import (
	"dns-hostlist-compiler/modules/app/cli"
	"dns-hostlist-compiler/modules/app/io"
	"dns-hostlist-compiler/modules/app/pipeline"
	"fmt"
	"log"
)

func main() {
	inputPath, outputPath := cli.ParseArgs()

	links, err := io.ReadLinksFromFile(inputPath)
	if err != nil {
		log.Fatalf("failed to read links: %v", err)
	}

	links = pipeline.DedupeSlice(links)

	rules, err := pipeline.RunPipeline(links)
	if err != nil {
		log.Fatalf("pipeline error: %v", err)
	}

	if err := io.WriteLines(outputPath, rules); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}

	fmt.Printf("Wrote %d rules to %s\n", len(rules), outputPath)
}
