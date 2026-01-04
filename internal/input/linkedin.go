package input

import (
	"fmt"
	"regexp"
	"strings"
)

// LinkedInProfile represents a parsed LinkedIn profile
type LinkedInProfile struct {
	Name       string
	Title      string
	Summary    string
	Experience []Experience
	Skills     []string
	Education  []Education
	RawProfile string
}

// Experience represents a work experience entry
type Experience struct {
	Title       string
	Company     string
	Duration    string
	Description string
}

// Education represents an education entry
type Education struct {
	School   string
	Degree   string
	Field    string
	Duration string
}

// LinkedInParser handles LinkedIn profile parsing
type LinkedInParser struct{}

// NewLinkedInParser creates a new LinkedIn parser
func NewLinkedInParser() *LinkedInParser {
	return &LinkedInParser{}
}

// ParseProfile parses a LinkedIn profile from fetched content
// This expects the HTML content or exported profile text
func (lp *LinkedInParser) ParseProfile(content string) (*LinkedInProfile, error) {
	if content == "" {
		return nil, fmt.Errorf("empty profile content")
	}

	profile := &LinkedInProfile{
		RawProfile: content,
		Experience: []Experience{},
		Skills:     []string{},
		Education:  []Education{},
	}

	// Extract name (usually at the beginning)
	profile.Name = extractName(content)

	// Extract job title
	profile.Title = extractTitle(content)

	// Extract summary
	profile.Summary = extractSummary(content)

	// Extract experience
	profile.Experience = extractExperience(content)

	// Extract skills
	profile.Skills = extractSkills(content)

	// Extract education
	profile.Education = extractEducation(content)

	return profile, nil
}

// ToText converts the profile to readable text format
func (p *LinkedInProfile) ToText() string {
	var sb strings.Builder

	if p.Name != "" {
		sb.WriteString("Name: " + p.Name + "\n")
	}

	if p.Title != "" {
		sb.WriteString("Current Title: " + p.Title + "\n")
	}

	if p.Summary != "" {
		sb.WriteString("\nSummary:\n" + p.Summary + "\n")
	}

	if len(p.Experience) > 0 {
		sb.WriteString("\nExperience:\n")
		for _, exp := range p.Experience {
			if exp.Title != "" {
				sb.WriteString("- " + exp.Title)
				if exp.Company != "" {
					sb.WriteString(" at " + exp.Company)
				}
				sb.WriteString("\n")
			}
			if exp.Duration != "" {
				sb.WriteString("  " + exp.Duration + "\n")
			}
			if exp.Description != "" {
				sb.WriteString("  " + exp.Description + "\n")
			}
		}
	}

	if len(p.Skills) > 0 {
		sb.WriteString("\nSkills:\n")
		for _, skill := range p.Skills {
			sb.WriteString("- " + skill + "\n")
		}
	}

	if len(p.Education) > 0 {
		sb.WriteString("\nEducation:\n")
		for _, edu := range p.Education {
			if edu.School != "" {
				sb.WriteString("- " + edu.School + "\n")
			}
			if edu.Degree != "" && edu.Field != "" {
				sb.WriteString("  " + edu.Degree + " in " + edu.Field + "\n")
			} else if edu.Degree != "" {
				sb.WriteString("  " + edu.Degree + "\n")
			}
			if edu.Duration != "" {
				sb.WriteString("  " + edu.Duration + "\n")
			}
		}
	}

	return sb.String()
}

// extractName extracts the profile name
func extractName(content string) string {
	// Look for common LinkedIn name patterns
	patterns := []string{
		`(?i)(?:name|profile)[\s:]*([A-Z][a-z]+\s+[A-Z][a-z]+)`,
		`^\s*([A-Z][a-z]+\s+[A-Z][a-z]+)`,
		`(?m)^\s*([A-Z][a-z]+\s+[A-Z][a-z]+)\s*$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractTitle extracts the job title
func extractTitle(content string) string {
	patterns := []string{
		`(?i)(?:title|position)[\s:]*([^\n]+)`,
		`(?i)(?:current position)[\s:]*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			if len(title) > 0 && len(title) < 100 {
				return title
			}
		}
	}

	// Fallback: Look for job title keywords on second line after name
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i > 0 && i < 5 { // Check first few lines after the name
			trimmed := strings.TrimSpace(line)
			if isLikelyJobTitle(trimmed) && len(trimmed) > 0 && len(trimmed) < 100 {
				return trimmed
			}
		}
	}

	return ""
}

// extractSummary extracts the profile summary
func extractSummary(content string) string {
	patterns := []string{
		`(?i)(?:summary|about)[\s:]*([^\n]+(?:\n[^\n]+){0,10})`,
		`(?i)(?:about me)[\s:]*([^\n]+(?:\n[^\n]+){0,10})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			summary := strings.TrimSpace(matches[1])
			if summary != "" {
				return summary
			}
		}
	}

	return ""
}

// extractExperience extracts work experience entries
func extractExperience(content string) []Experience {
	var experiences []Experience

	// Split content by common experience markers
	lines := strings.Split(content, "\n")
	var currentExp Experience
	inExpSection := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for experience section start
		if strings.Contains(strings.ToLower(line), "experience") {
			inExpSection = true
			continue
		}

		if !inExpSection {
			continue
		}

		// Look for job title (usually followed by company)
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])

			// Check if current line looks like a job title
			if isLikelyJobTitle(line) && (isLikelyCompanyName(nextLine) || isLikelyDuration(nextLine)) {
				if currentExp.Title != "" {
					experiences = append(experiences, currentExp)
				}

				currentExp = Experience{
					Title: line,
				}

				// Next line might be company
				if isLikelyCompanyName(nextLine) {
					currentExp.Company = nextLine
					i++
				}
			} else if isLikelyDuration(line) {
				currentExp.Duration = line
			} else if isLikelyDescription(line) && currentExp.Title != "" {
				if currentExp.Description != "" {
					currentExp.Description += " " + line
				} else {
					currentExp.Description = line
				}
			}
		}
	}

	if currentExp.Title != "" {
		experiences = append(experiences, currentExp)
	}

	return experiences
}

// extractSkills extracts skills from the profile
func extractSkills(content string) []string {
	var skills []string
	lines := strings.Split(content, "\n")
	inSkillsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(strings.ToLower(trimmed), "skills") {
			inSkillsSection = true
			continue
		}

		if inSkillsSection {
			// Stop at next section
			if strings.Contains(strings.ToLower(trimmed), "experience") ||
				strings.Contains(strings.ToLower(trimmed), "education") ||
				strings.Contains(strings.ToLower(trimmed), "summary") {
				break
			}

			// Extract individual skills
			if trimmed != "" && !strings.HasPrefix(trimmed, "Endorsements") {
				// Remove common prefixes
				skill := strings.TrimPrefix(trimmed, "- ")
				skill = strings.TrimSpace(skill)

				if len(skill) > 0 && len(skill) < 100 && !strings.Contains(skill, "\n") {
					skills = append(skills, skill)
				}
			}
		}
	}

	return skills
}

// extractEducation extracts education entries
func extractEducation(content string) []Education {
	var educations []Education
	lines := strings.Split(content, "\n")
	inEduSection := false
	var currentEdu Education

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(strings.ToLower(trimmed), "education") {
			inEduSection = true
			continue
		}

		if !inEduSection {
			continue
		}

		// Stop at next section
		if (strings.Contains(strings.ToLower(trimmed), "skills") ||
			strings.Contains(strings.ToLower(trimmed), "experience")) &&
			currentEdu.School != "" {
			educations = append(educations, currentEdu)
			break
		}

		if trimmed == "" {
			if currentEdu.School != "" {
				educations = append(educations, currentEdu)
				currentEdu = Education{}
			}
			continue
		}

		// School name (usually capitalized)
		if isLikelySchoolName(trimmed) {
			if currentEdu.School != "" {
				educations = append(educations, currentEdu)
			}
			currentEdu = Education{School: trimmed}
		} else if currentEdu.School != "" {
			// Degree or field
			if isLikelyDegree(trimmed) {
				currentEdu.Degree = trimmed
			} else if isLikelyDuration(trimmed) {
				currentEdu.Duration = trimmed
			} else {
				currentEdu.Field = trimmed
			}
		}
	}

	if currentEdu.School != "" {
		educations = append(educations, currentEdu)
	}

	return educations
}

// Helper functions to identify content types

func isLikelyJobTitle(s string) bool {
	// Common job title keywords
	keywords := []string{
		"engineer", "developer", "manager", "architect", "analyst",
		"consultant", "designer", "scientist", "director", "senior",
		"junior", "lead", "associate", "specialist", "coordinator",
		"officer", "executive", "administrator", "technician",
	}

	lower := strings.ToLower(s)
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
}

func isLikelyCompanyName(s string) bool {
	// Company names typically start with capital letters
	return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' && len(s) < 100
}

func isLikelyDuration(s string) bool {
	patterns := []string{
		"20[0-9]{2}", // Years
		"jan", "feb", "mar", "apr", "may", "jun",
		"jul", "aug", "sep", "oct", "nov", "dec",
		"present", "current", "-",
	}

	lower := strings.ToLower(s)
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

func isLikelyDescription(s string) bool {
	return len(s) > 10 && len(s) < 500
}

func isLikelySchoolName(s string) bool {
	// University names typically contain specific keywords
	keywords := []string{
		"university", "college", "institute", "school", "academy",
		"polytechnic", "technical", "state", "central",
	}

	lower := strings.ToLower(s)
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
}

func isLikelyDegree(s string) bool {
	keywords := []string{
		"bachelor", "master", "phd", "diploma", "certificate",
		"associate", "b.a", "b.s", "m.a", "m.s", "m.b.a",
		"b.tech", "m.tech",
	}

	lower := strings.ToLower(s)
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
}
