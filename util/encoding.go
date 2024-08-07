package util

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func ConvertFileToANSI(inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	reader := transform.NewReader(inputFile, charmap.Windows1251.NewEncoder())

	_, err = io.Copy(outputFile, reader)
	if err != nil {
		return fmt.Errorf("error copying data to output file: %v", err)
	}

	return nil
}
