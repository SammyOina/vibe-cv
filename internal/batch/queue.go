package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sammyoina/vibe-cv/internal/db"
	"github.com/sammyoina/vibe-cv/internal/llm"
)

// JobQueue manages async batch jobs
type JobQueue struct {
	mu       sync.RWMutex
	repo     *db.Repository
	provider llm.Provider
	workers  int
	jobChan  chan int
	stopChan chan struct{}
}

// NewJobQueue creates a new job queue
func NewJobQueue(repo *db.Repository, workers int) *JobQueue {
	return &JobQueue{
		repo:     repo,
		workers:  workers,
		jobChan:  make(chan int, 100),
		stopChan: make(chan struct{}),
	}
}

// SetProvider sets the LLM provider for batch job processing
func (q *JobQueue) SetProvider(provider llm.Provider) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.provider = provider
}

// Start starts the job queue workers
func (q *JobQueue) Start() {
	for i := 0; i < q.workers; i++ {
		go q.worker()
	}
}

// Stop stops the job queue workers
func (q *JobQueue) Stop() {
	close(q.stopChan)
}

// worker processes jobs from the queue
func (q *JobQueue) worker() {
	for {
		select {
		case <-q.stopChan:
			return
		case jobID := <-q.jobChan:
			q.ProcessJob(jobID)
		}
	}
}

// CreateJob creates a new batch job with items
func (q *JobQueue) CreateJob(identityID *int, totalItems int) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, err := q.repo.CreateBatchJob(identityID, totalItems)
	if err != nil {
		return 0, err
	}

	// Queue the job for processing
	select {
	case q.jobChan <- job.ID:
	default:
		// Queue full, skip for now
	}

	return job.ID, nil
}

// AddJobItem adds an item to a batch job
func (q *JobQueue) AddJobItem(batchJobID int, cvID *int, jobDescription string) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.repo.CreateBatchJobItem(batchJobID, cvID, jobDescription)
	if err != nil {
		return 0, err
	}

	return item.ID, nil
}

// ProcessJob processes a specific batch job with actual LLM calls
func (q *JobQueue) ProcessJob(jobID int) error {
	q.mu.Lock()

	// Update job status to processing
	if err := q.repo.UpdateBatchJobStatus(jobID, "processing", 0); err != nil {
		q.mu.Unlock()
		return err
	}

	// Check if provider is set
	if q.provider == nil {
		q.mu.Unlock()
		q.repo.UpdateBatchJobStatus(jobID, "failed", 0)
		return fmt.Errorf("LLM provider not set for batch processing")
	}

	q.mu.Unlock()

	// Get job items
	items, err := q.repo.GetBatchJobItems(jobID)
	if err != nil {
		q.repo.UpdateBatchJobStatus(jobID, "failed", 0)
		return err
	}

	// Process each item with actual LLM calls
	completedCount := 0
	ctx := context.Background()
	
	for _, item := range items {
		// Get CV text for this item
		var cvText string
		if item.CVID != nil {
			cv, err := q.repo.GetCV(*item.CVID)
			if err != nil {
				// Skip this item on error
				q.repo.UpdateBatchJobItem(item.ID, "failed", nil, nil)
				continue
			}
			cvText = cv.OriginalText
		}

		// Call LLM provider to customize CV
		result, err := q.provider.Customize(ctx, cvText, item.JobDescription, []string{})
		if err != nil {
			// Mark item as failed
			q.repo.UpdateBatchJobItem(item.ID, "failed", nil, nil)
			continue
		}

		// Store successful result
		resultData := map[string]interface{}{
			"status":          "completed",
			"match_score":     result.MatchScore,
			"modifications":   result.Modifications,
			"customized_text": result.ModifiedCV,
			"processed_at":    time.Now().UTC(),
		}
		resultJSON, _ := json.Marshal(resultData)
		resultPtr := (*json.RawMessage)(&resultJSON)

		if err := q.repo.UpdateBatchJobItem(item.ID, "completed", resultPtr, nil); err != nil {
			q.repo.UpdateBatchJobStatus(jobID, "failed", completedCount)
			return err
		}
		completedCount++
	}

	// Mark job as completed
	return q.repo.UpdateBatchJobStatus(jobID, "completed", completedCount)
}

// GetBatchJobStatus retrieves the status of a batch job
func (q *JobQueue) GetBatchJobStatus(jobID int) (map[string]interface{}, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, err := q.repo.GetBatchJob(jobID)
	if err != nil {
		return nil, err
	}

	items, err := q.repo.GetBatchJobItems(jobID)
	if err != nil {
		return nil, err
	}

	// Calculate progress
	completedCount := 0
	for _, item := range items {
		if item.Status == "completed" {
			completedCount++
		}
	}

	progress := 0.0
	if len(items) > 0 {
		progress = float64(completedCount) / float64(len(items))
	}

	return map[string]interface{}{
		"job_id":     job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
		"total":      len(items),
		"completed":  completedCount,
		"progress":   progress,
	}, nil
}
