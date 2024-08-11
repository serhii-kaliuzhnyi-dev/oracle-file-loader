package convertor

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/db"
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

func ConvertFileToUtf8(inputPath, outputPath string) error {
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

	reader := transform.NewReader(inputFile, charmap.Windows1251.NewDecoder())
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		return fmt.Errorf("error converting file to UTF-8: %v", err)
	}

	return nil
}

func FilterConvertedFile(filePath string, tableConfig *db.TableConfig, delimiter rune) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening converted file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = delimiter
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading converted file: %v", err)
	}

	headers := records[0]
	var filteredHeaders []string
	var filteredRecords [][]string

	// Identify columns to keep based on tableConfig
	keepColumns := make(map[int]bool)
	for i, header := range headers {
		for _, colInfo := range tableConfig.Columns {
			if colInfo.OriginalName == header && colInfo.Create {
				keepColumns[i] = true
				filteredHeaders = append(filteredHeaders, header)
				break
			}
		}
	}

	// Filter records based on keepColumns
	for _, record := range records[1:] {
		var filteredRecord []string
		for i, value := range record {
			if keepColumns[i] {
				filteredRecord = append(filteredRecord, value)
			}
		}
		filteredRecords = append(filteredRecords, filteredRecord)
	}

	// Write the filtered records to a new converted file
	filteredFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating filtered file: %v", err)
	}
	defer filteredFile.Close()

	writer := csv.NewWriter(filteredFile)
	writer.Comma = delimiter

	// Write the headers
	if err := writer.Write(filteredHeaders); err != nil {
		return fmt.Errorf("error writing headers to filtered file: %v", err)
	}

	// Write the filtered records
	for _, record := range filteredRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing records to filtered file: %v", err)
		}
	}

	writer.Flush()
	return writer.Error()
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
