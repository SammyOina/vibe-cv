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
func (lg *LaTeXGenerator) GeneratePDF(cvContent string, filename string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(lg.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate LaTeX template
	latexContent := generateLaTeXTemplate(cvContent)

	// Write LaTeX file
	texFile := filepath.Join(lg.outputDir, filename+".tex")

	err := os.WriteFile(texFile, []byte(latexContent), 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to write LaTeX file: %w", err)
	}

	// Compile LaTeX to PDF
	cmd := exec.Command(lg.laTeXPath,
		"-interaction=nonstopmode",
		"-output-directory="+lg.outputDir,
		texFile)

	// Capture both stdout and stderr for better error reporting
	output, err := cmd.CombinedOutput()

	// Return path to generated PDF
	pdfFile := filepath.Join(lg.outputDir, filename+".pdf")

	// Check if PDF was actually created, even if pdflatex returned errors
	// pdflatex often returns non-zero for warnings/non-fatal errors but still produces output
	if _, statErr := os.Stat(pdfFile); statErr == nil {
		return pdfFile, nil
	}

	// PDF was not created - return the actual error
	if err != nil {
		return "", fmt.Errorf("failed to compile LaTeX: %w\nLaTeX output:\n%s", err, string(output))
	}

	return "", fmt.Errorf("PDF file was not created at %s\nLaTeX output:\n%s", pdfFile, string(output))
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
