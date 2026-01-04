package input

import (
	"regexp"
	"strings"
)

// StructuredCVContent represents parsed CV content with structure
type StructuredCVContent struct {
	RawText        string
	Name           string
	Title          string
	Email          string
	Phone          string
	Summary        string
	Experience     []Experience
	Skills         []string
	Education      []Education
	Certifications []string
}

// StructuredJobDescription represents parsed job description with structure
type StructuredJobDescription struct {
	RawText          string
	Title            string
	Company          string
	Location         string
	Description      string
	Requirements     []string
	Responsibilities []string
	PreferredSkills  []string
	BenefitsKeywords []string
}

// EnhancedParser provides structured parsing of CVs and job descriptions
type EnhancedParser struct{}

// NewEnhancedParser creates a new enhanced parser
func NewEnhancedParser() *EnhancedParser {
	return &EnhancedParser{}
}

// ParseCV extracts structured information from CV text
func (ep *EnhancedParser) ParseCV(content string) *StructuredCVContent {
	cv := &StructuredCVContent{
		RawText:        content,
		Experience:     []Experience{},
		Skills:         []string{},
		Education:      []Education{},
		Certifications: []string{},
	}

	// Extract contact information
	cv.Email = extractEmail(content)
	cv.Phone = extractPhone(content)

	// Extract name (usually at top)
	cv.Name = extractName(content)

	// Extract professional title
	cv.Title = extractTitle(content)

	// Extract summary/objective
	cv.Summary = extractSummary(content)

	// Extract sections
	cv.Experience = ep.extractExperienceSection(content)
	cv.Skills = ep.extractSkillsSection(content)
	cv.Education = ep.extractEducationSection(content)
	cv.Certifications = ep.extractCertificationsSection(content)

	// If no title found, use the first job title from experience
	if cv.Title == "" && len(cv.Experience) > 0 {
		cv.Title = cv.Experience[0].Title
	}

	return cv
}

// ParseJobDescription extracts structured information from job description text
func (ep *EnhancedParser) ParseJobDescription(content string) *StructuredJobDescription {
	job := &StructuredJobDescription{
		RawText:          content,
		Requirements:     []string{},
		Responsibilities: []string{},
		PreferredSkills:  []string{},
		BenefitsKeywords: []string{},
	}

	// Extract job title
	job.Title = ep.extractJobTitle(content)

	// Extract company name
	job.Company = ep.extractCompanyName(content)

	// Extract location
	job.Location = ep.extractLocation(content)

	// Extract job description sections
	job.Description = ep.extractDescription(content)
	job.Requirements = ep.extractRequirements(content)
	job.Responsibilities = ep.extractResponsibilities(content)
	job.PreferredSkills = ep.extractPreferredSkills(content)
	job.BenefitsKeywords = ep.extractBenefitsKeywords(content)

	return job
}

// extractEmail extracts email address from text
func extractEmail(content string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	matches := emailRegex.FindAllString(content, -1)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// extractPhone extracts phone number from text
func extractPhone(content string) string {
	phoneRegex := regexp.MustCompile(`(?:\+\d{1,3}[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
	matches := phoneRegex.FindAllString(content, -1)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// extractExperienceSection extracts work experience from CV
func (ep *EnhancedParser) extractExperienceSection(content string) []Experience {
	var experiences []Experience

	// Find experience section
	expStart := findSectionStart(content, []string{"experience", "work experience", "employment"})
	if expStart == -1 {
		return experiences
	}

	expEnd := findSectionEnd(content, expStart, []string{"education", "skills", "certification", "projects"})
	if expEnd == -1 {
		expEnd = len(content)
	}

	sectionContent := content[expStart:expEnd]

	// Parse individual entries
	entries := parseBulletedList(sectionContent)
	for _, entry := range entries {
		exp := Experience{
			Title:       extractJobTitleFromEntry(entry),
			Company:     extractCompanyFromEntry(entry),
			Duration:    extractDurationFromEntry(entry),
			Description: extractDescriptionFromEntry(entry),
		}
		if exp.Title != "" {
			experiences = append(experiences, exp)
		}
	}

	return experiences
}

// extractEducationSection extracts education from CV
func (ep *EnhancedParser) extractEducationSection(content string) []Education {
	var educations []Education

	eduStart := findSectionStart(content, []string{"education", "academic"})
	if eduStart == -1 {
		return educations
	}

	eduEnd := findSectionEnd(content, eduStart, []string{"experience", "skills", "certification"})
	if eduEnd == -1 {
		eduEnd = len(content)
	}

	sectionContent := content[eduStart:eduEnd]
	entries := parseBulletedList(sectionContent)

	for _, entry := range entries {
		edu := Education{
			School:   extractSchoolFromEntry(entry),
			Degree:   extractDegreeFromEntry(entry),
			Field:    extractFieldFromEntry(entry),
			Duration: extractDurationFromEntry(entry),
		}
		if edu.School != "" {
			educations = append(educations, edu)
		}
	}

	return educations
}

// extractSkillsSection extracts skills from CV
func (ep *EnhancedParser) extractSkillsSection(content string) []string {
	var skills []string

	skillStart := findSectionStart(content, []string{"skills", "technical skills", "competencies"})
	if skillStart == -1 {
		return skills
	}

	skillEnd := findSectionEnd(content, skillStart, []string{"experience", "education", "projects"})
	if skillEnd == -1 {
		skillEnd = len(content)
	}

	sectionContent := content[skillStart:skillEnd]

	// Parse comma or bullet separated skills
	entries := strings.FieldsFunc(sectionContent, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})

	for _, entry := range entries {
		skill := strings.TrimSpace(entry)
		skill = strings.TrimPrefix(skill, "-")
		skill = strings.TrimSpace(skill)

		if len(skill) > 0 && len(skill) < 100 {
			skills = append(skills, skill)
		}
	}

	return skills
}

// extractCertificationsSection extracts certifications from CV
func (ep *EnhancedParser) extractCertificationsSection(content string) []string {
	var certs []string

	certStart := findSectionStart(content, []string{"certification", "certifications", "licenses"})
	if certStart == -1 {
		return certs
	}

	certEnd := findSectionEnd(content, certStart, []string{"skills", "projects"})
	if certEnd == -1 {
		certEnd = len(content)
	}

	sectionContent := content[certStart:certEnd]
	entries := parseBulletedList(sectionContent)

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry != "" {
			certs = append(certs, entry)
		}
	}

	return certs
}

// extractJobTitle extracts job title from job description
func (ep *EnhancedParser) extractJobTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines[:min(10, len(lines))] {
		line = strings.TrimSpace(line)
		if isLikelyJobTitle(line) && len(line) < 100 {
			return line
		}
	}
	return ""
}

// extractCompanyName extracts company name from job description
func (ep *EnhancedParser) extractCompanyName(content string) string {
	patterns := []string{
		`(?i)company[\s:]*([^\n]+)`,
		`(?i)employer[\s:]*([^\n]+)`,
		`(?i)posted by[\s:]*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			company := strings.TrimSpace(matches[1])
			if company != "" && len(company) < 200 {
				return company
			}
		}
	}

	return ""
}

// extractLocation extracts location from job description
func (ep *EnhancedParser) extractLocation(content string) string {
	patterns := []string{
		`(?i)location[\s:]*([^\n]+)`,
		`(?i)based in[\s:]*([^\n]+)`,
		`(?i)city[\s:]*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			location := strings.TrimSpace(matches[1])
			if location != "" && len(location) < 200 {
				return location
			}
		}
	}

	return ""
}

// extractDescription extracts job description overview
func (ep *EnhancedParser) extractDescription(content string) string {
	// Get first few paragraphs before Requirements section
	descStart := 0
	descEnd := findSectionStart(content, []string{"requirements", "responsibilities", "qualifications"})
	if descEnd == -1 {
		descEnd = min(len(content), 500)
	}

	desc := strings.TrimSpace(content[descStart:descEnd])
	// Remove job title if it's at the beginning
	lines := strings.Split(desc, "\n")
	if len(lines) > 0 && isLikelyJobTitle(lines[0]) {
		lines = lines[1:]
	}

	return strings.Join(lines, "\n")
}

// extractRequirements extracts job requirements
func (ep *EnhancedParser) extractRequirements(content string) []string {
	return ep.extractListSection(content, []string{"requirements", "must have", "required"})
}

// extractResponsibilities extracts job responsibilities
func (ep *EnhancedParser) extractResponsibilities(content string) []string {
	return ep.extractListSection(content, []string{"responsibilities", "your role", "what you'll do"})
}

// extractPreferredSkills extracts preferred/nice-to-have skills
func (ep *EnhancedParser) extractPreferredSkills(content string) []string {
	return ep.extractListSection(content, []string{"preferred", "nice to have", "bonus"})
}

// extractBenefitsKeywords extracts benefits mentioned
func (ep *EnhancedParser) extractBenefitsKeywords(content string) []string {
	return ep.extractListSection(content, []string{"benefits", "perks", "compensation"})
}

// extractListSection extracts items from a bulleted/numbered list section
func (ep *EnhancedParser) extractListSection(content string, sectionNames []string) []string {
	var items []string

	start := findSectionStart(content, sectionNames)
	if start == -1 {
		return items
	}

	end := findSectionEnd(content, start, []string{"education", "company", "location", "salary"})
	if end == -1 {
		end = len(content)
	}

	sectionContent := content[start:end]
	entries := parseBulletedList(sectionContent)

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry != "" && len(entry) < 500 {
			items = append(items, entry)
		}
	}

	return items
}

// Helper functions

func findSectionStart(content string, sectionNames []string) int {
	lower := strings.ToLower(content)
	for _, name := range sectionNames {
		idx := strings.Index(lower, strings.ToLower(name))
		if idx != -1 {
			// Return index after the section name
			return idx + len(name)
		}
	}
	return -1
}

func findSectionEnd(content string, start int, nextSectionNames []string) int {
	lower := strings.ToLower(content[start:])
	for _, name := range nextSectionNames {
		idx := strings.Index(lower, strings.ToLower(name))
		if idx != -1 {
			return start + idx
		}
	}
	return -1
}

func parseBulletedList(content string) []string {
	var items []string
	lines := strings.Split(content, "\n")
	var currentItem string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a bullet point or number
		isBullet := strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "•") ||
			strings.HasPrefix(trimmed, "*") ||
			regexp.MustCompile(`^\d+[\.\)]\s`).MatchString(trimmed)

		if isBullet {
			if currentItem != "" {
				items = append(items, strings.TrimSpace(currentItem))
			}
			// Remove bullet markers
			currentItem = regexp.MustCompile(`^[-•*]+\s*|\d+[\.\)]\s*`).ReplaceAllString(trimmed, "")
		} else if trimmed != "" && currentItem != "" {
			// Continuation of previous item
			currentItem += " " + trimmed
		} else if trimmed != "" {
			currentItem = trimmed
		}
	}

	if currentItem != "" {
		items = append(items, strings.TrimSpace(currentItem))
	}

	return items
}

func extractJobTitleFromEntry(entry string) string {
	lines := strings.Split(entry, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

func extractCompanyFromEntry(entry string) string {
	lines := strings.Split(entry, "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return ""
}

func extractDurationFromEntry(entry string) string {
	re := regexp.MustCompile(`(20[0-9]{2}|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec|present)`)
	matches := re.FindStringSubmatch(strings.ToLower(entry))
	if len(matches) > 0 {
		// Find the full duration string
		lower := strings.ToLower(entry)
		idx := strings.Index(lower, strings.ToLower(matches[0]))
		if idx != -1 {
			end := min(idx+30, len(entry))
			return strings.TrimSpace(entry[idx:end])
		}
	}
	return ""
}

func extractDescriptionFromEntry(entry string) string {
	lines := strings.Split(entry, "\n")
	if len(lines) > 2 {
		return strings.Join(lines[2:], " ")
	}
	return ""
}

func extractSchoolFromEntry(entry string) string {
	lines := strings.Split(entry, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

func extractDegreeFromEntry(entry string) string {
	keywords := []string{"bachelor", "master", "phd", "diploma", "associate", "degree"}
	lower := strings.ToLower(entry)

	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			// Extract the line containing the keyword
			lines := strings.Split(entry, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), keyword) {
					return strings.TrimSpace(line)
				}
			}
		}
	}
	return ""
}

func extractFieldFromEntry(entry string) string {
	lines := strings.Split(entry, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "20") && len(strings.TrimSpace(line)) > 3 {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
