// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package input

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// DOCXParser handles DOCX file parsing.
type DOCXParser struct {
	maxFileSize int64
}

// NewDOCXParser creates a new DOCX parser.
func NewDOCXParser(maxFileSize int64) *DOCXParser {
	if maxFileSize == 0 {
		maxFileSize = 50 * 1024 * 1024 // 50MB default
	}

	return &DOCXParser{
		maxFileSize: maxFileSize,
	}
}

// ParseFile extracts text from a DOCX file.
func (p *DOCXParser) ParseFile(filePath string) (string, error) {
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

// ParseBytes extracts text from DOCX bytes.
func (p *DOCXParser) ParseBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty DOCX data")
	}

	// DOCX is a ZIP file, try to open it as such
	reader, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid DOCX file: %w", err)
	}

	// Look for document.xml
	var docFile *zip.File

	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			docFile = f

			break
		}
	}

	if docFile == nil {
		return "", errors.New("document.xml not found in DOCX file")
	}

	// Read and parse document.xml
	rc, err := docFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("failed to read document.xml: %w", err)
	}

	// Parse XML and extract text
	text := extractTextFromWordML(string(content))

	if text == "" {
		return "", errors.New("no text content found in DOCX")
	}

	return text, nil
}

// WordMLDocument represents the structure of document.xml.
type WordMLDocument struct {
	Body struct {
		Paragraphs []struct {
			Text []struct {
				Content string `xml:",innerxml"`
			} `xml:"w:r/w:t"`
		} `xml:"w:p"`
	} `xml:"w:body"`
}

// extractTextFromWordML extracts text from Word ML XML.
func extractTextFromWordML(xmlContent string) string {
	var doc WordMLDocument

	// Parse XML
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))

	err := decoder.Decode(&doc)
	if err != nil && err != io.EOF {
		// If XML parsing fails, try simple regex extraction
		return extractTextFromRawXML(xmlContent)
	}

	var result strings.Builder

	for _, para := range doc.Body.Paragraphs {
		for _, t := range para.Text {
			result.WriteString(t.Content)
		}

		result.WriteRune('\n')
	}

	text := result.String()

	// Clean up whitespace
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

// extractTextFromRawXML extracts text from XML using simple pattern matching.
func extractTextFromRawXML(xmlContent string) string {
	var result strings.Builder

	// Extract text within <w:t> tags
	parts := strings.Split(xmlContent, "<w:t")
	for i := 1; i < len(parts); i++ {
		// Find the content between > and </w:t>
		start := strings.Index(parts[i], ">")
		end := strings.Index(parts[i], "</w:t>")

		if start != -1 && end != -1 && start < end {
			text := parts[i][start+1 : end]
			result.WriteString(text)
		}
	}

	// Replace XML entities
	text := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&apos;", "'",
	).Replace(result.String())

	// Clean up whitespace
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
