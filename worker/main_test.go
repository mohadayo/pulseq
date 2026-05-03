package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "healthy" {
		t.Fatalf("expected healthy, got %v", resp["status"])
	}
	if resp["service"] != "worker" {
		t.Fatalf("expected worker, got %v", resp["service"])
	}
}

func TestProcessHandler(t *testing.T) {
	mu.Lock()
	jobs = nil
	stats = WorkerStats{}
	mu.Unlock()

	job := Job{ID: "test-1", QueueName: "q1", Body: "hello"}
	body, _ := json.Marshal(job)
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp Job
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "processed" {
		t.Fatalf("expected processed, got %s", resp.Status)
	}
	if resp.ProcessedAt == 0 {
		t.Fatal("expected ProcessedAt to be set")
	}
}

func TestProcessHandlerInvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/process", nil)
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestProcessHandlerMissingFields(t *testing.T) {
	job := Job{ID: "", Body: ""}
	body, _ := json.Marshal(job)
	req := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()
	processHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestStatsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()
	statsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp WorkerStats
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Uptime <= 0 {
		t.Fatal("expected positive uptime")
	}
}

func TestJobsHandler(t *testing.T) {
	mu.Lock()
	jobs = []Job{{ID: "j1", Body: "test", Status: "processed"}}
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	w := httptest.NewRecorder()
	jobsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []Job
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 job, got %d", len(resp))
	}
}
