package db

import (
    "database/sql"
    "errors"
    "fmt"
    "sync"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

// DB is a global database handle protected by a mutex. SQLite
// connections are safe for concurrent use as long as multithreading
// mode is enabled (the default). The mutex is used for schema
// initialization only.
var (
    database *sql.DB
    once     sync.Once
)

// Init opens the SQLite database at the given path and creates the
// necessary schema if it doesn't already exist.
func Init(path string) error {
    var err error
    once.Do(func() {
        database, err = sql.Open("sqlite3", path)
        if err != nil {
            return
        }
        // Set busy timeout to avoid SQLITE_BUSY errors
        database.Exec("PRAGMA busy_timeout = 5000")
        schema := []string{
            // jobs table stores job metadata
            `CREATE TABLE IF NOT EXISTS jobs (
                id TEXT PRIMARY KEY,
                command TEXT,
                start_time DATETIME,
                end_time DATETIME,
                timeout INTEGER,
                status TEXT
            );`,
            // hosts table stores hosts associated with a job
            `CREATE TABLE IF NOT EXISTS hosts (
                job_id TEXT,
                host TEXT
            );`,
            // results table stores the output per host
            `CREATE TABLE IF NOT EXISTS results (
                job_id TEXT,
                host TEXT,
                output TEXT,
                error TEXT,
                exit_code INTEGER,
                duration INTEGER
            );`,
        }
        for _, stmt := range schema {
            if _, err = database.Exec(stmt); err != nil {
                return
            }
        }
    })
    return err
}

// Job represents a job summary.
type Job struct {
    ID        string    `json:"id"`
    Command   string    `json:"command"`
    StartTime time.Time `json:"start_time"`
    EndTime   *time.Time `json:"end_time"`
    Timeout   int       `json:"timeout"`
    Status    string    `json:"status"`
}

// Result represents an individual host execution result.
type Result struct {
    Host     string `json:"host"`
    Output   string `json:"output"`
    Error    string `json:"error"`
    ExitCode int    `json:"exit_code"`
    Duration int64  `json:"duration"`
}

// CreateJob inserts a new job row.
func CreateJob(id, command string, startTime time.Time, timeout int, status string) error {
    if database == nil {
        return errors.New("database not initialized")
    }
    _, err := database.Exec(`INSERT INTO jobs (id, command, start_time, timeout, status) VALUES (?, ?, ?, ?, ?)`, id, command, startTime, timeout, status)
    return err
}

// InsertHost associates a host with a job.
func InsertHost(id, host string) error {
    if database == nil {
        return errors.New("database not initialized")
    }
    _, err := database.Exec(`INSERT INTO hosts (job_id, host) VALUES (?, ?)`, id, host)
    return err
}

// UpdateJobStatus updates the status and end_time of a job.
func UpdateJobStatus(id, status string, endTime time.Time) error {
    if database == nil {
        return errors.New("database not initialized")
    }
    _, err := database.Exec(`UPDATE jobs SET status = ?, end_time = ? WHERE id = ?`, status, endTime, id)
    return err
}

// InsertResult inserts a result row.
func InsertResult(id, host, output, errMsg string, exitCode int, duration int64) error {
    if database == nil {
        return errors.New("database not initialized")
    }
    _, err := database.Exec(`INSERT INTO results (job_id, host, output, error, exit_code, duration) VALUES (?, ?, ?, ?, ?, ?)`, id, host, output, errMsg, exitCode, duration)
    return err
}

// GetJobs returns a slice of job summaries. If there are no jobs, an empty slice is returned.
func GetJobs() ([]Job, error) {
    if database == nil {
        return nil, errors.New("database not initialized")
    }
    rows, err := database.Query(`SELECT id, command, start_time, end_time, timeout, status FROM jobs ORDER BY start_time DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var jobs []Job
    for rows.Next() {
        var job Job
        var endTime sql.NullTime
        if err := rows.Scan(&job.ID, &job.Command, &job.StartTime, &endTime, &job.Timeout, &job.Status); err != nil {
            return nil, err
        }
        if endTime.Valid {
            t := endTime.Time
            job.EndTime = &t
        }
        jobs = append(jobs, job)
    }
    return jobs, rows.Err()
}

// GetJob retrieves a job by id and its associated results.
func GetJob(id string) (Job, []Result, error) {
    if database == nil {
        return Job{}, nil, errors.New("database not initialized")
    }
    var job Job
    var endTime sql.NullTime
    row := database.QueryRow(`SELECT id, command, start_time, end_time, timeout, status FROM jobs WHERE id = ?`, id)
    if err := row.Scan(&job.ID, &job.Command, &job.StartTime, &endTime, &job.Timeout, &job.Status); err != nil {
        if err == sql.ErrNoRows {
            return Job{}, nil, fmt.Errorf("job %s not found", id)
        }
        return Job{}, nil, err
    }
    if endTime.Valid {
        t := endTime.Time
        job.EndTime = &t
    }
    // Fetch results
    rows, err := database.Query(`SELECT host, output, error, exit_code, duration FROM results WHERE job_id = ?`, id)
    if err != nil {
        return job, nil, err
    }
    defer rows.Close()
    var results []Result
    for rows.Next() {
        var r Result
        if err := rows.Scan(&r.Host, &r.Output, &r.Error, &r.ExitCode, &r.Duration); err != nil {
            return job, nil, err
        }
        results = append(results, r)
    }
    return job, results, rows.Err()
}
