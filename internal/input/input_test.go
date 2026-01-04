package input

import (
	"strings"
	"testing"
)

func TestFetcher_ExtractHTML(t *testing.T) {
	html := `<html><body><h1>Job Title</h1><p>This is a test job description</p></body></html>`
	result := extractTextFromHTML(html)

	if !strings.Contains(result, "Job Title") {
		t.Errorf("Expected 'Job Title' in result, got: %s", result)
	}

	if !strings.Contains(result, "test job description") {
		t.Errorf("Expected 'test job description' in result, got: %s", result)
	}
}

func TestPDFParser_ValidatePDF(t *testing.T) {
	parser := NewPDFParser(0)

	// Invalid PDF
	invalidData := []byte("not a pdf")
	_, err := parser.ParseBytes(invalidData)
	if err == nil {
		t.Errorf("Expected error for invalid PDF")
	}
}

func TestDOCXParser_ZipValidation(t *testing.T) {
	parser := NewDOCXParser(0)

	// Invalid DOCX
	invalidData := []byte("not a docx")
	_, err := parser.ParseBytes(invalidData)
	if err == nil {
		t.Errorf("Expected error for invalid DOCX")
	}
}

func TestEnhancedParser_ExtractEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Contact: john.doe@example.com", "john.doe@example.com"},
		{"Email: test+tag@domain.co.uk", "test+tag@domain.co.uk"},
		{"No email here", ""},
	}

	for _, tt := range tests {
		result := extractEmail(tt.input)
		if result != tt.expected {
			t.Errorf("extractEmail(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestEnhancedParser_ExtractPhone(t *testing.T) {
	tests := []struct {
		input    string
		hasPhone bool
	}{
		{"Call me at (123) 456-7890", true},
		{"Phone: +1-234-567-8900", true},
		{"No phone", false},
	}

	for _, tt := range tests {
		result := extractPhone(tt.input)
		hasResult := result != ""
		if hasResult != tt.hasPhone {
			t.Errorf("extractPhone(%s) found phone = %v, want %v", tt.input, hasResult, tt.hasPhone)
		}
	}
}

func TestLinkedInParser_ParseProfile(t *testing.T) {
	profileText := `
	John Doe
	Senior Software Engineer
	
	Summary:
	Experienced software engineer with 5+ years of experience
	
	Experience:
	- Senior Engineer at Tech Corp
	  2020 - Present
	  Led development of cloud infrastructure
	
	Skills:
	- Go
	- Kubernetes
	- AWS
	
	Education:
	- University of Tech
	  BS in Computer Science
	  2015 - 2019
	`

	parser := NewLinkedInParser()
	profile, err := parser.ParseProfile(profileText)

	if err != nil {
		t.Errorf("ParseProfile returned error: %v", err)
	}

	if profile.Name == "" {
		t.Error("Expected name to be extracted")
	}

	if profile.Title == "" {
		t.Error("Expected title to be extracted")
	}

	if len(profile.Skills) == 0 {
		t.Error("Expected skills to be extracted")
	}

	if len(profile.Experience) == 0 {
		t.Error("Expected experience to be extracted")
	}

	if len(profile.Education) == 0 {
		t.Error("Expected education to be extracted")
	}
}

func TestEnhancedParser_ParseCV(t *testing.T) {
	cvText := `
	John Doe
	john@example.com
	(123) 456-7890
	
	Professional Summary:
	Software engineer with 10 years of experience
	
	Work Experience:
	Senior Software Engineer
	Acme Corp
	2020 - Present
	Led development of microservices architecture
	
	Skills:
	Go, Python, Kubernetes, AWS, Docker
	
	Education:
	BS in Computer Science
	University of State
	2012 - 2016
	`

	parser := NewEnhancedParser()
	cv := parser.ParseCV(cvText)

	if cv.Name == "" {
		t.Error("Expected name to be extracted")
	}

	if cv.Email == "" {
		t.Error("Expected email to be extracted")
	}

	if cv.Phone == "" {
		t.Error("Expected phone to be extracted")
	}

	if cv.Title == "" {
		t.Error("Expected title to be extracted")
	}

	if len(cv.Skills) == 0 {
		t.Error("Expected skills to be extracted")
	}

	if len(cv.Experience) == 0 {
		t.Error("Expected experience to be extracted")
	}

	if len(cv.Education) == 0 {
		t.Error("Expected education to be extracted")
	}
}

func TestEnhancedParser_ParseJobDescription(t *testing.T) {
	jobText := `
	Senior Go Engineer
	
	Company: Tech Startup
	Location: San Francisco, CA
	
	We are looking for an experienced Go engineer.
	
	Requirements:
	- 5+ years of Go experience
	- Experience with Kubernetes
	- Experience with microservices
	
	Responsibilities:
	- Design and implement APIs
	- Mentor junior developers
	- Code reviews
	
	Preferred Skills:
	- Experience with gRPC
	- Experience with Docker
	
	Benefits:
	- Competitive salary
	- Health insurance
	- Remote work
	`

	parser := NewEnhancedParser()
	job := parser.ParseJobDescription(jobText)

	if job.Title == "" {
		t.Error("Expected job title to be extracted")
	}

	if job.Company == "" {
		t.Error("Expected company to be extracted")
	}

	if job.Location == "" {
		t.Error("Expected location to be extracted")
	}

	if len(job.Requirements) == 0 {
		t.Error("Expected requirements to be extracted")
	}

	if len(job.Responsibilities) == 0 {
		t.Error("Expected responsibilities to be extracted")
	}

	if len(job.PreferredSkills) == 0 {
		t.Error("Expected preferred skills to be extracted")
	}

	if len(job.BenefitsKeywords) == 0 {
		t.Error("Expected benefits to be extracted")
	}
}

func TestIsValidJobURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://www.linkedin.com/jobs/view/12345", true},
		{"https://indeed.com/jobs?q=golang", true},
		{"https://company.com/careers", true},
		{"https://example.com/random-page", false},
		{"https://example.com", false},
	}

	for _, tt := range tests {
		result := IsValidJobURL(tt.url)
		if result != tt.valid {
			t.Errorf("IsValidJobURL(%s) = %v, want %v", tt.url, result, tt.valid)
		}
	}
}
