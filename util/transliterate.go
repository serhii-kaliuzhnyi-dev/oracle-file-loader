package util

import (
	"strconv"

	"github.com/mozillazg/go-unidecode"
)

func TransliterateHeaders(headers []string) []string {
	transliterated := make([]string, len(headers))
	emptyCount := 1
	for i, header := range headers {
		if header == "" {
			transliterated[i] = "empty" + strconv.Itoa(emptyCount)
			emptyCount++
		} else {
			transliterated[i] = unidecode.Unidecode(header)
		}
	}
	return transliterated
}
