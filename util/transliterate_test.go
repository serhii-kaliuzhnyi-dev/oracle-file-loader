package util

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func TestTransliterateHeaders_UTF8(t *testing.T) {
	headers := []string{"Привет", "Мир", "Тест"}
	expected := []string{"Privet", "Mir", "Test"}

	result := TransliterateHeaders(headers)

	for i, header := range result {
		if header != expected[i] {
			t.Errorf("Expected %s but got %s", expected[i], header)
		}
	}
}

func TestTransliterateHeaders_Windows1251(t *testing.T) {
	headers := []string{"Привет", "Мир", "Тест"}
	expected := []string{"Privet", "Mir", "Test"}

	// Encode headers to Windows-1251
	var windows1251Headers []string
	for _, header := range headers {
		encodedHeader, err := encodeWindows1251(header)
		if err != nil {
			t.Fatalf("Failed to encode header to Windows-1251: %v", err)
		}
		windows1251Headers = append(windows1251Headers, encodedHeader)
	}

	// Decode headers back to UTF-8 for transliteration
	var utf8Headers []string
	for _, header := range windows1251Headers {
		decodedHeader, err := decodeWindows1251(header)
		if err != nil {
			t.Fatalf("Failed to decode header from Windows-1251: %v", err)
		}
		utf8Headers = append(utf8Headers, decodedHeader)
	}

	result := TransliterateHeaders(utf8Headers)

	for i, header := range result {
		if header != expected[i] {
			t.Errorf("Expected %s but got %s", expected[i], header)
		}
	}
}

func TestTransliterateHeaders_Empty(t *testing.T) {
	headers := []string{"", ""}
	expected := []string{"empty1", "empty2"}

	result := TransliterateHeaders(headers)

	for i, header := range result {
		if header != expected[i] {
			t.Errorf("Expected %s but got %s", expected[i], header)
		}
	}
}

func TestTransliterateHeaders_Mixed(t *testing.T) {
	headers := []string{"Привет", "", "Тест", ""}
	expected := []string{"Privet", "empty1", "Test", "empty2"}

	result := TransliterateHeaders(headers)

	for i, header := range result {
		if header != expected[i] {
			t.Errorf("Expected %s but got %s", expected[i], header)
		}
	}
}

func encodeWindows1251(s string) (string, error) {
	encoder := charmap.Windows1251.NewEncoder()
	reader := transform.NewReader(strings.NewReader(s), encoder)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func decodeWindows1251(s string) (string, error) {
	decoder := charmap.Windows1251.NewDecoder()
	reader := transform.NewReader(strings.NewReader(s), decoder)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
