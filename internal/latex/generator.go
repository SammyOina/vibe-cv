// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LaTeXGenerator generates LaTeX content and compiles to PDF.
type LaTeXGenerator struct {
	outputDir string
	laTeXPath string
}

// NewLaTeXGenerator creates a new LaTeX generator.
func NewLaTeXGenerator(outputDir, laTeXPath string) *LaTeXGenerator {
	return &LaTeXGenerator{
		outputDir: outputDir,
		laTeXPath: laTeXPath,
	}
}

// GeneratePDF generates a PDF from CV content.
func (lg *LaTeXGenerator) GeneratePDF(cvContent string, filename string, isFullLatex bool) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(lg.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	var latexContent string
	if isFullLatex {
		latexContent = cvContent // Use LLM's generated LaTeX directly
	} else {
		latexContent = generateLaTeXTemplate(cvContent) // Old fallback
	}

	// AGGRESSIVE EXTRACTION: Strip all markdown fences and conversational text.
	docStart := strings.Index(latexContent, "\\documentclass")
	docEnd := strings.Index(latexContent, "\\end{document}")
	if docStart != -1 && docEnd != -1 {
		latexContent = latexContent[docStart : docEnd+len("\\end{document}")]
		
		// CLEANUP: Strip trailing manual line breaks before the end of the document
		// This prevents the "\\\end{document}" error.
		latexContent = strings.ReplaceAll(latexContent, "\\\\\n\\end{document}", "\n\\end{document}")
		latexContent = strings.ReplaceAll(latexContent, "\\\\ \n\\end{document}", "\n\\end{document}")
	} else if docStart != -1 {
		latexContent = latexContent[docStart:]
	}

	// Write LaTeX file
	texFile := filepath.Join(lg.outputDir, filename+".tex")
	err := os.WriteFile(texFile, []byte(latexContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write LaTeX file: %w", err)
	}

	// Run pdflatex twice to resolve cross-references (hyperref bookmarks, \ref, \pageref)
	var output []byte
	for i := 0; i < 2; i++ {
		cmd := exec.Command(lg.laTeXPath,
			"-interaction=nonstopmode",
			"-output-directory="+lg.outputDir,
			texFile)
		output, err = cmd.CombinedOutput()
	}

	// Return path to generated PDF
	pdfFile := filepath.Join(lg.outputDir, filename+".pdf")

	// Helper to extract log contents safely
	getLogContents := func() string {
		logFile := filepath.Join(lg.outputDir, filename+".log")
		if logContent, readErr := os.ReadFile(logFile); readErr == nil {
			return string(logContent)
		}
		return string(output)
	}

	// Check if PDF was actually created, even if pdflatex returned errors
	// pdflatex often returns non-zero for warnings/non-fatal errors but still produces output
	if _, statErr := os.Stat(pdfFile); statErr == nil {
		// Check if the log file contains critical errors (not just warnings)
		logStr := getLogContents()
		// These indicate the PDF is likely incomplete/broken
		if strings.Contains(logStr, "Emergency stop") ||
			strings.Contains(logStr, "Fatal error") || 
			strings.Contains(logStr, "!</error>") || 
			strings.Contains(logStr, "Undefined control sequence") {
			
			excerptLen := 1000
			if len(logStr) < excerptLen {
				excerptLen = len(logStr)
			}
			return "", fmt.Errorf("COMPILATION_ERROR: LaTeX compilation produced a broken PDF: critical errors in log\nLog excerpt:\n%s", logStr[len(logStr)-excerptLen:])
		}
		
		return pdfFile, nil
	}

	// PDF was not created - return the actual error alongside the compiler log
	logStr := getLogContents()
	excerptLen := 1000
	if len(logStr) < excerptLen {
		excerptLen = len(logStr)
	}

	if err != nil {
		return "", fmt.Errorf("COMPILATION_ERROR: failed to compile LaTeX: %w\nLaTeX output/Log:\n%s", err, logStr[len(logStr)-excerptLen:])
	}

	return "", fmt.Errorf("COMPILATION_ERROR: PDF file was not created at %s\nLaTeX output/Log:\n%s", pdfFile, logStr[len(logStr)-excerptLen:])
}

// generateLaTeXTemplate generates a basic LaTeX template for CV.
func generateLaTeXTemplate(cvContent string) string {
	escaped := escapeLatex(cvContent)
	template := `\documentclass[11pt,a4paper]{article}
\usepackage[utf8]{inputenc}
\usepackage[margin=0.5in]{geometry}
\usepackage{hyperref}
\usepackage{xcolor}
\usepackage{fancyhdr}

\pagestyle{fancy}
\fancyhf{}
\renewcommand{\headrulewidth}{0pt}

\setlength{\parindent}{0pt}
\setlength{\parskip}{0.5em}

\title{Curriculum Vitae}
\author{}
\date{}

\begin{document}

\section*{Professional Summary}
` + escaped + `

\end{document}`

	return template
}

func escapeLatex(s string) string {
	replacer := strings.NewReplacer(
		"&", `\&`,
		"%", `\%`,
		"$", `\$`,
		"#", `\#`,
		"_", `\_`,
		"{", `\{`,
		"}", `\}`,
		"~", `\textasciitilde{}`,
		"^", `\textasciicircum{}`,
	)

	return replacer.Replace(s)
}
