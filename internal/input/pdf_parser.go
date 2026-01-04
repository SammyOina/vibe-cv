// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package input

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// PDFParser handles PDF file parsing.
type PDFParser struct {
	maxFileSize int64
}

// NewPDFParser creates a new PDF parser.
func NewPDFParser(maxFileSize int64) *PDFParser {
	if maxFileSize == 0 {
		maxFileSize = 50 * 1024 * 1024 // 50MB default
	}

	return &PDFParser{
		maxFileSize: maxFileSize,
	}
}

// ParseFile extracts text from a PDF file.
func (p *PDFParser) ParseFile(filePath string) (string, error) {
	// Check file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() > p.maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed (%d bytes)", p.maxFileSize)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read all content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseBytes(content)
}

// ParseBytes extracts text from PDF bytes
// Note: For full PDF parsing, consider using:
// - github.com/unidoc/unipdf
// - github.com/pdfcpu/pdfcpu.
func (p *PDFParser) ParseBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty PDF data")
	}

	// Check PDF signature
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return "", errors.New("not a valid PDF file")
	}

	// For MVP Phase 2, extract basic text content
	// This is a simplified implementation
	text := extractTextFromPDF(data)

	if text == "" {
		return "", errors.New("no text content found in PDF")
	}

	return text, nil
}

// extractTextFromPDF performs basic text extraction from PDF
// This is a simplified version - for production use a dedicated library.
func extractTextFromPDF(data []byte) string {
	// Try to extract text between BT (begin text) and ET (end text) operators
	var result strings.Builder

	inText := false
	i := 0

	for i < len(data) {
		// Look for "BT" operator
		if i+1 < len(data) && data[i] == 'B' && data[i+1] == 'T' {
			inText = true
			i += 2

			continue
		}

		// Look for "ET" operator
		if i+1 < len(data) && data[i] == 'E' && data[i+1] == 'T' {
			inText = false

			result.WriteRune('\n')

			i += 2

			continue
		}

		if inText {
			// Look for text strings in parentheses or angle brackets
			if data[i] == '(' {
				j := i + 1
				for j < len(data) && data[j] != ')' {
					if data[j] >= 32 && data[j] < 127 {
						result.WriteByte(data[j])
					}

					j++
				}

				i = j

				continue
			}
		}

		i++
	}

	// Clean up whitespace
	text := result.String()
	lines := strings.Split(text, "\n")

	var cleanedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}
