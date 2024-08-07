package convertor

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

	// Detect file encoding
	encoding, err := detectFileEncoding(inputFile)
	if err != nil {
		return fmt.Errorf("error detecting file encoding: %v", err)
	}
	_, err = inputFile.Seek(0, io.SeekStart) // Reset file pointer
	if err != nil {
		return fmt.Errorf("error resetting file pointer: %v", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	var reader io.Reader
	if encoding == "windows-1251" {
		// If the file is already in windows-1251, just copy it directly
		reader = inputFile
	} else {
		// Assume UTF-8 and convert to windows-1251
		reader = transform.NewReader(inputFile, charmap.Windows1251.NewEncoder())
	}

	_, err = io.Copy(outputFile, reader)
	if err != nil {
		return fmt.Errorf("error copying data to output file: %v", err)
	}

	return nil
}

func detectFileEncoding(file *os.File) (string, error) {
	buffer := make([]byte, 1024)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Basic heuristic to detect windows-1251 encoding
	for _, b := range buffer {
		if b >= 0x80 && b <= 0xFF {
			return "windows-1251", nil
		}
	}

	return "utf-8", nil
}
