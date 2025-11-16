package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		specFile   string
		outputFile string
	)

	flag.StringVar(&specFile, "spec", "", "Path to the YAML specification file")
	flag.StringVar(&outputFile, "output", "", "Path to the output Go file")
	flag.Parse()

	if specFile == "" || outputFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -spec <yaml-file> -output <go-file>\n", os.Args[0])
		os.Exit(1)
	}

	// Load the specification
	spec, err := LoadSpec(specFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading spec: %v\n", err)
		os.Exit(1)
	}

	// Generate the code
	generator := NewGenerator(spec)
	if err := generator.WriteToFile(outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s from %s\n", outputFile, specFile)
}
