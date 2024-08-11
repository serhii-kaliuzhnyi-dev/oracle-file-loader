package util

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GenerateConvertedFilePath(filePath string) string {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), ext)
	return filepath.Join(filepath.Dir(filePath), base+"--converted"+ext)
}

func GenerateUtf8FilePath(filePath string) string {
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)
	return filepath.Join(filepath.Dir(filePath), name+"_utf8"+ext)
}

func DetectDelimiter(filePath string) (rune, error) {
	file, err := os.Open(filePath)
	
	if err != nil {
		return 0, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	line, _, err := reader.ReadLine()

	if err != nil {
		return 0, err
	}

	semicolonCount := strings.Count(string(line), ";")
	tabCount := strings.Count(string(line), "\t")
	commaCount := strings.Count(string(line), ",")

	if semicolonCount > tabCount && semicolonCount > commaCount {
		return ';', nil
	} else if tabCount > semicolonCount && tabCount > commaCount {
		return '\t', nil
	} else if commaCount > semicolonCount && commaCount > tabCount {
		return ',', nil
	}

	return 0, fmt.Errorf("unable to detect delimiter")
}

func IsNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}