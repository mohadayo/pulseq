package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Job struct {
	ID          string  `json:"id"`
	QueueName   string  `json:"queue_name"`
	Body        string  `json:"body"`
	Status      string  `json:"status"`
	ProcessedAt float64 `json:"processed_at,omitempty"`
}

type WorkerStats struct {
	TotalProcessed int     `json:"total_processed"`
	TotalFailed    int     `json:"total_failed"`
	Uptime         float64 `json:"uptime_seconds"`
}

var (
	mu        sync.Mutex
	jobs      []Job
	stats     WorkerStats
	startTime time.Time
)

func init() {
	startTime = time.Now()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":    "healthy",
		"service":   "worker",
		"timestamp": time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		log.Printf("[ERROR] Failed to decode job: %v", err)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if job.ID == "" || job.Body == "" {
		http.Error(w, `{"error":"id and body are required"}`, http.StatusBadRequest)
		return
	}

	job.Status = "processed"
	job.ProcessedAt = float64(time.Now().UnixMilli()) / 1000.0

	mu.Lock()
	jobs = append(jobs, job)
	stats.TotalProcessed++
	mu.Unlock()

	log.Printf("[INFO] Processed job %s from queue '%s'", job.ID, job.QueueName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	s := WorkerStats{
		TotalProcessed: stats.TotalProcessed,
		TotalFailed:    stats.TotalFailed,
		Uptime:         time.Since(startTime).Seconds(),
	}
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	j := make([]Job, len(jobs))
	copy(j, jobs)
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

func main() {
	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "5001"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/process", processHandler)
	mux.HandleFunc("/stats", statsHandler)
	mux.HandleFunc("/jobs", jobsHandler)

	log.Printf("[INFO] Starting PulseQ Worker on port %s", port)
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}
}
