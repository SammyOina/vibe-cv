// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"fmt"
	"regexp"
	"strings"
)

// Packages that are known to be unavailable and have safe replacements.
var packageReplacements = map[string]string{
	"lato":          "helvet",
	"fontawesome5":  "", // remove — rarely actually used
	"fontawesome":   "", // remove
	"roboto":        "helvet",
	"opensans":      "helvet",
	"sourcesanspro": "helvet",
	"raleway":       "helvet",
	"inter":         "helvet",
}

// dangerousCommands are LaTeX commands that could execute arbitrary code.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\\write18\b`),
	regexp.MustCompile(`\\immediate\s*\\write18\b`),
	regexp.MustCompile(`\\input\s*\{[/~]`),  // absolute/home path includes
	regexp.MustCompile(`\\include\s*\{[/~]`), // absolute/home path includes
	regexp.MustCompile(`\\openin\b`),
	regexp.MustCompile(`\\openout\b`),
	regexp.MustCompile(`\\newwrite\b`),
	regexp.MustCompile(`\\catcode`),
}

// markdownFencePattern matches ``` code fences that LLMs sometimes include.
var markdownFencePattern = regexp.MustCompile("(?m)^\\s*```(?:latex|tex)?\\s*$")

// usepackagePattern matches \usepackage[options]{name} lines.
var usepackagePattern = regexp.MustCompile(`\\usepackage(?:\[([^\]]*)\])?\{([^}]+)\}`)

// environmentPattern matches \begin{env} and \end{env}.
var beginEnvPattern = regexp.MustCompile(`\\begin\{([^}]+)\}`)
var endEnvPattern = regexp.MustCompile(`\\end\{([^}]+)\}`)

// SanitizeLatex cleans LLM-generated LaTeX content to maximize compilation success.
// It strips dangerous commands, replaces unavailable packages, removes markdown
// artifacts, and validates basic document structure.
// Returns sanitized content and nil error if fixable, or error if unfixable.
func SanitizeLatex(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("empty LaTeX content")
	}

	// Step 1: Strip markdown code fences (LLMs often wrap LaTeX in ```latex ... ```)
	content = stripMarkdownFences(content)

	// Step 2: Strip dangerous commands
	content = stripDangerousCommands(content)

	// Step 3: Replace unavailable packages with safe alternatives
	content = replaceUnavailablePackages(content)

	// Step 4: Escape unprotected & characters in text-mode content
	content = escapeUnprotectedAmpersands(content)

	// Step 5: Validate basic document structure
	if err := validateStructure(content); err != nil {
		return content, fmt.Errorf("invalid LaTeX structure: %w", err)
	}

	// Step 6: Validate balanced environments
	if err := validateEnvironments(content); err != nil {
		return content, fmt.Errorf("unbalanced LaTeX environments: %w", err)
	}

	return content, nil
}

// stripMarkdownFences removes markdown code fences that LLMs sometimes include
// around LaTeX content.
func stripMarkdownFences(content string) string {
	// Remove lines that are just ``` or ```latex or ```tex
	content = markdownFencePattern.ReplaceAllString(content, "")

	// Trim leading/trailing whitespace that may remain
	content = strings.TrimSpace(content)

	return content
}

// stripDangerousCommands removes LaTeX commands that could execute arbitrary code
// or access the filesystem.
func stripDangerousCommands(content string) string {
	for _, pattern := range dangerousPatterns {
		content = pattern.ReplaceAllString(content, "% [REMOVED: dangerous command]")
	}
	return content
}

// replaceUnavailablePackages substitutes packages that aren't in the Docker image
// with safe alternatives.
func replaceUnavailablePackages(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a \usepackage line
		matches := usepackagePattern.FindStringSubmatch(trimmed)
		if len(matches) >= 3 {
			pkgName := strings.TrimSpace(matches[2])

			// Handle comma-separated package lists
			if strings.Contains(pkgName, ",") {
				pkgs := strings.Split(pkgName, ",")
				var keptPkgs []string
				for _, p := range pkgs {
					p = strings.TrimSpace(p)
					if replacement, needsReplace := packageReplacements[p]; needsReplace {
						if replacement != "" {
							keptPkgs = append(keptPkgs, replacement)
						}
						// if replacement is "", we drop the package
					} else {
						keptPkgs = append(keptPkgs, p)
					}
				}
				if len(keptPkgs) == 0 {
					result = append(result, "% "+line+" % [REPLACED: unavailable packages]")
					continue
				}
				opts := ""
				if matches[1] != "" {
					opts = "[" + matches[1] + "]"
				}
				result = append(result, `\usepackage`+opts+`{`+strings.Join(keptPkgs, ",")+`}`)
				continue
			}

			// Single package
			if replacement, needsReplace := packageReplacements[pkgName]; needsReplace {
				if replacement == "" {
					result = append(result, "% "+line+" % [REPLACED: unavailable package]")
				} else {
					opts := ""
					if matches[1] != "" {
						opts = "[" + matches[1] + "]"
					}
					result = append(result, `\usepackage`+opts+`{`+replacement+`}`)
				}
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// escapeUnprotectedAmpersands escapes bare & characters in text-mode lines.
// It replaces & not already preceded by \ with \&, while skipping lines
// inside tabular/math environments where & is a valid alignment character.
func escapeUnprotectedAmpersands(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	// Environments where & is a valid alignment tab, not a text character.
	tabularEnvs := map[string]bool{
		"tabular": true, "tabularx": true, "tabular*": true,
		"array": true, "longtable": true, "longtabu": true,
		"tabu": true, "align": true, "align*": true,
		"aligned": true, "eqnarray": true, "eqnarray*": true,
		"matrix": true, "pmatrix": true, "bmatrix": true,
		"vmatrix": true, "Vmatrix": true,
	}

	inTabularEnv := false
	tabularEnvDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track entry/exit of alignment environments.
		for _, m := range beginEnvPattern.FindAllStringSubmatch(line, -1) {
			if tabularEnvs[m[1]] {
				inTabularEnv = true
				tabularEnvDepth++
			}
		}
		for _, m := range endEnvPattern.FindAllStringSubmatch(line, -1) {
			if tabularEnvs[m[1]] {
				tabularEnvDepth--
				if tabularEnvDepth <= 0 {
					inTabularEnv = false
					tabularEnvDepth = 0
				}
			}
		}

		// Inside alignment environments & is meaningful — leave it alone.
		if inTabularEnv {
			result = append(result, line)
			continue
		}

		// Pure LaTeX command lines and comment lines are unlikely to hold
		// stray & characters, so skip them to avoid touching things like
		// \newcommand definitions that happen to contain &.
		if strings.HasPrefix(trimmed, "\\") || strings.HasPrefix(trimmed, "%") {
			result = append(result, line)
			continue
		}

		// Walk the line rune-by-rune so we can check the preceding character
		// and avoid double-escaping already-escaped \& sequences.
		var escaped strings.Builder
		runes := []rune(line)
		for i, r := range runes {
			if r == '&' && (i == 0 || runes[i-1] != '\\') {
				escaped.WriteString(`\&`)
			} else {
				escaped.WriteRune(r)
			}
		}
		result = append(result, escaped.String())
	}

	return strings.Join(result, "\n")
}

// validateStructure checks that the LaTeX has the minimum required structure.
func validateStructure(content string) error {
	if !strings.Contains(content, `\documentclass`) {
		return fmt.Errorf("missing \\documentclass declaration")
	}
	if !strings.Contains(content, `\begin{document}`) {
		return fmt.Errorf("missing \\begin{document}")
	}
	if !strings.Contains(content, `\end{document}`) {
		return fmt.Errorf("missing \\end{document}")
	}

	// Check that \begin{document} comes after \documentclass
	docclassIdx := strings.Index(content, `\documentclass`)
	beginDocIdx := strings.Index(content, `\begin{document}`)
	if beginDocIdx < docclassIdx {
		return fmt.Errorf("\\begin{document} appears before \\documentclass")
	}

	return nil
}

// validateEnvironments checks that all \begin{env} have matching \end{env}.
func validateEnvironments(content string) error {
	begins := beginEnvPattern.FindAllStringSubmatch(content, -1)
	ends := endEnvPattern.FindAllStringSubmatch(content, -1)

	envCount := make(map[string]int)
	for _, match := range begins {
		envCount[match[1]]++
	}
	for _, match := range ends {
		envCount[match[1]]--
	}

	var unbalanced []string
	for env, count := range envCount {
		if count != 0 {
			unbalanced = append(unbalanced, fmt.Sprintf("%s (opens: %d more than closes)", env, count))
		}
	}

	if len(unbalanced) > 0 {
		return fmt.Errorf("unbalanced environments: %s", strings.Join(unbalanced, ", "))
	}

	return nil
}
