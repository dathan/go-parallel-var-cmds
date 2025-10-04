package db

import (
    "os"
    "path/filepath"
    "reflect"
    "testing"
    "time"

    "github.com/google/uuid"
)

// TestJobLifecycle ensures that jobs can be created, updated and retrieved along
// with their hosts and results.
func TestJobLifecycle(t *testing.T) {
    tmp := t.TempDir()
    dbPath := filepath.Join(tmp, "test.db")
    if err := Init(dbPath); err != nil {
        t.Fatalf("Init() error = %v", err)
    }
    id := uuid.New().String()
    start := time.Now().UTC()
    if err := CreateJob(id, "echo hello", start, 10, "running"); err != nil {
        t.Fatalf("CreateJob() error = %v", err)
    }
    // insert hosts
    hosts := []string{"hostA", "hostB"}
    for _, h := range hosts {
        if err := InsertHost(id, h); err != nil {
            t.Fatalf("InsertHost() error = %v", err)
        }
    }
    // insert result
    if err := InsertResult(id, "all", "output", "", 0, 100); err != nil {
        t.Fatalf("InsertResult() error = %v", err)
    }
    // update job
    end := start.Add(5 * time.Second)
    if err := UpdateJobStatus(id, "completed", end); err != nil {
        t.Fatalf("UpdateJobStatus() error = %v", err)
    }
    // fetch job summary list
    jobs, err := GetJobs()
    if err != nil {
        t.Fatalf("GetJobs() error = %v", err)
    }
    if len(jobs) != 1 {
        t.Fatalf("expected 1 job, got %d", len(jobs))
    }
    // fetch job with results
    job, results, err := GetJob(id)
    if err != nil {
        t.Fatalf("GetJob() error = %v", err)
    }
    if job.ID != id || job.Command != "echo hello" || job.Status != "completed" {
        t.Fatalf("unexpected job %+v", job)
    }
    // we expect one result row
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    r := results[0]
    if !reflect.DeepEqual(r, Result{Host: "all", Output: "output", Error: "", ExitCode: 0, Duration: 100}) {
        t.Fatalf("unexpected result %+v", r)
    }
    // ensure temp database file exists
    if _, err := os.Stat(dbPath); err != nil {
        t.Fatalf("database file was not created: %v", err)
    }
}
