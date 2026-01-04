# vibe-cv

> AI-powered CV customization tool that tailors your resume to specific job opportunities

## Overview

**vibe-cv** is an intelligent resume customization platform that automatically adapts your CV to match job descriptions. Powered by LLMs, it analyzes job requirements and contextual information to create tailored, optimized CVs in PDF format.

## Features

### Core Functionality

- **Intelligent CV Customization**: Uses LLMs to analyze job descriptions and customize your CV accordingly
- **Multiple Input Sources**: 
  - Job descriptions (text input)
  - Job links (auto-extraction of job details)
  - Additional context via text input or document uploads
  - LinkedIn profile links
  - Document uploads (PDF, DOCX, etc.)
  
- **LaTeX-based PDF Generation**: Creates professionally formatted CVs by:
  - Building optimized LaTeX templates
  - Compiling to high-quality PDF output
  
- **Agentic Flow**: Advanced workflow capabilities that can:
  - Break down complex customization tasks
  - Iteratively refine CV content
  - Validate and optimize output
  - Adapt strategy based on job requirements

- **Flexible LLM Configuration**: 
  - Support for multiple LLM providers
  - User/admin configurable LLM selection
  - Easy integration of new LLM backends

## Architecture

### Technology Stack

- **Primary Language**: Go
- **Additional Languages**: Other languages may be used as needed for specialized tasks
- **Output Format**: LaTeX → PDF
- **LLM Integration**: Pluggable architecture supporting multiple providers

### High-Level Components

```
┌─────────────────────────────────────┐
│   Input Processing                  │
│  (Job Description, Links, Docs)    │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Context Extraction & Analysis     │
│  (Job Requirements, Keywords)       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   LLM-Powered Customization Engine  │
│  (Agentic workflows, optimization) │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   LaTeX Template Generation         │
│  (CV content formatting)            │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   PDF Compilation & Output          │
│  (Final resume generation)          │
└─────────────────────────────────────┘
```

## Getting Started

### Prerequisites

- Go 1.25.0 or later
- LaTeX installation (for PDF compilation)
  - **Ubuntu/Debian**: `sudo apt-get install texlive-latex-base texlive-latex-extra`
  - **macOS**: `brew install mactex` or download from [MacTeX](https://www.tug.org/mactex/)
  - **Windows**: Download from [MiKTeX](https://miktex.org/) or [TeX Live](https://www.tug.org/texlive/)
- API keys for LLM providers (OpenAI, Anthropic, or Google Gemini)
- PostgreSQL (optional - for persistence features)

### Installation

```bash
git clone https://github.com/SammyOina/vibe-cv.git
cd vibe-cv
go mod download
```

### Configuration

The application can be configured using environment variables or a `.env` file in the project root.

#### Required Environment Variables

```env
# LLM Configuration (Required)
LLM_API_KEY=your_api_key_here
```

#### Optional Environment Variables

```env
# LLM Provider Settings
LLM_PROVIDER=openai          # Options: openai, anthropic, gemini (default: openai)
LLM_MODEL=gpt-4              # Model to use (default: gpt-4)

# Server Configuration
SERVER_HOST=localhost        # Server host (default: localhost)
SERVER_PORT=8080            # Server port (default: 8080)

# Output Configuration
OUTPUT_DIR=./outputs        # Directory for generated PDFs (default: ./outputs)
LATEX_PATH=pdflatex        # Path to LaTeX compiler (default: pdflatex)

# Database Configuration (Optional - enables persistence features)
DATABASE_URL=postgres://user:password@localhost:5432/vibecv

# Authentication Configuration (Optional - Ory Kratos integration)
KRATOS_ENABLED=false                           # Enable Kratos auth (default: false)
KRATOS_PUBLIC_URL=http://localhost:4433       # Kratos public API URL
KRATOS_ADMIN_URL=http://localhost:4434        # Kratos admin API URL
```

### Docker Setup

The project includes Docker Compose configuration for easy deployment with PostgreSQL and Ory Kratos authentication:

```bash
cd docker
cp .env.example .env  # Configure your environment variables
docker-compose up -d
```

This will start:
- PostgreSQL database
- Ory Kratos identity server
- vibe-cv application server

### Usage

1. Set up your environment variables (see Configuration section above)
2. Run the application:

```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080` by default (configurable via `SERVER_HOST` and `SERVER_PORT`).

## Quickstart

### 1. Basic CV Customization with Job Description Text

Customize your CV by providing raw CV content and job description as text:

```bash
curl -X POST http://localhost:8080/api/latest/customize-cv \
  -H "Content-Type: application/json" \
  -d '{
    "cv": "Your CV content here...",
    "job_description": "We are looking for a Senior Backend Engineer with 5+ years of Go experience, expertise in cloud technologies, and strong system design knowledge.",
    "llm_config": {
      "provider": "openai",
      "model": "gpt-4"
    }
  }'
```

Response:
```json
{
  "status": "success",
  "customized_cv_url": "/outputs/cv-1.pdf",
  "match_score": 0.85,
  "modifications": ["Added cloud technologies section", "Highlighted Go experience"]
}
```

### 2. Customize CV with Additional Context

Add supplementary information for better customization:

```bash
curl -X POST http://localhost:8080/api/latest/customize-cv \
  -H "Content-Type: application/json" \
  -d '{
    "cv": "Your CV content here...",
    "job_description": "Senior Backend Engineer - Go, Kubernetes, AWS",
    "additional_context": [
      {
        "type": "text",
        "content": "Specialized in microservices architecture and distributed systems"
      },
      {
        "type": "text",
        "content": "5+ years of production Go experience with high-traffic systems"
      }
    ],
    "llm_config": {
      "provider": "openai",
      "model": "gpt-4"
    }
  }'
```

### 3. Use Anthropic Claude for Customization

```bash
curl -X POST http://localhost:8080/api/latest/customize-cv \
  -H "Content-Type: application/json" \
  -d '{
    "cv": "Your CV content here...",
    "job_description": "Full-stack developer with React and Node.js experience",
    "llm_config": {
      "provider": "anthropic",
      "model": "claude-3-opus"
    }
  }'
```

### 4. Use Google Gemini for Customization

```bash
curl -X POST http://localhost:8080/api/latest/customize-cv \
  -H "Content-Type: application/json" \
  -d '{
    "cv": "Your CV content here...",
    "job_description": "Data Science Engineer with ML expertise",
    "llm_config": {
      "provider": "gemini",
      "model": "gemini-pro"
    }
  }'
```

### 5. Batch Customization (Multiple Job Applications)

Customize your CV for multiple positions in a single batch:

```bash
curl -X POST http://localhost:8080/api/latest/batch-customize \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "cv": "Your CV content here...",
        "job_description": "Backend Engineer - Python, Django, PostgreSQL"
      },
      {
        "cv": "Your CV content here...",
        "job_description": "DevOps Engineer - Kubernetes, Terraform, CI/CD"
      }
    ]
  }'
```

Response:
```json
{
  "job_id": "batch-123",
  "status": "processing"
}
```

### 6. Download Customized CV as PDF

Download a customized CV version as a PDF file. The version ID is returned from the customize-cv response:

```bash
curl -X GET http://localhost:8080/api/latest/download/2 -o my-customized-cv.pdf
```

This will retrieve the CV from the database and generate a PDF with proper formatting.

### 7. Check Batch Job Status

```bash
curl -X GET http://localhost:8080/api/latest/batch/batch-123/status
```

### 8. Retrieve CV Versions

Get all versions created for a specific CV:

```bash
curl -X GET http://localhost:8080/api/latest/versions/1
```

### 9. Get Version Details

Retrieve detailed information about a specific CV version:

```bash
curl -X GET http://localhost:8080/api/latest/versions/1/detail
```

### 10. Compare Two CV Versions

Compare modifications between two versions:

```bash
curl -X POST http://localhost:8080/api/latest/compare-versions \
  -H "Content-Type: application/json" \
  -d '{
    "version_id_1": 1,
    "version_id_2": 2
  }'
```

## API Design

### Customize CV

```bash
POST /api/latest/customize-cv
```

**Request:**

```json
{
  "cv": "Raw CV content as text",
  "job_description": "Job description text",
  "additional_context": [
    {
      "type": "text",
      "content": "Additional skills or achievements"
    }
  ],
  "llm_config": {
    "provider": "openai",
    "model": "gpt-4"
  }
}
```

**Response:**

```json
{
  "status": "success",
  "customized_cv_url": "/outputs/cv-1.pdf",
  "match_score": 0.92,
  "modifications": [
    "Highlighted relevant experience",
    "Added specific keywords from job description",
    "Reordered skills based on job requirements"
  ]
}
```

### Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/latest/customize-cv` | Customize CV for a job description |
| `POST` | `/api/latest/batch-customize` | Submit batch customization jobs |
| `GET` | `/api/latest/versions/{cv_id}` | List all versions for a CV |
| `GET` | `/api/latest/versions/{version_id}/detail` | Get detailed version info |
| `GET` | `/api/latest/download/{version_id}` | Download customized CV as PDF |
| `POST` | `/api/latest/compare-versions` | Compare two CV versions |
| `GET` | `/api/latest/analytics` | Get user analytics |
| `GET` | `/api/latest/dashboard` | Get global dashboard stats |
| `GET` | `/api/latest/batch/{job_id}/status` | Check batch job status |
| `GET` | `/api/latest/batch/{job_id}/download` | Download batch results |
| `GET` | `/api/latest/health` | Health check endpoint |

## LLM Integration

The project uses a pluggable LLM architecture that supports:

- **OpenAI**: GPT-4, GPT-3.5-turbo
- **Anthropic**: Claude 3 series
- **Google**: Gemini
- **Local Models**: Ollama, LM Studio, etc.
- **Future**: Azure OpenAI, Vertex AI, etc.

Configuration is flexible and can be set per-request or globally.

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see LICENSE file for details.

## Contact

- GitHub: [@SammyOina](https://github.com/SammyOina)
- Issues: [GitHub Issues](https://github.com/SammyOina/vibe-cv/issues)

## Acknowledgments

- Built with Go for performance and simplicity
- Leveraging modern LLM capabilities for intelligent CV customization
- Inspired by the need to make job applications more efficient