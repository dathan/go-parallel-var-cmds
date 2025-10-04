package executor

import (
    "bufio"
    "context"
    "errors"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/dathan/go-parallel-var-cmds/internal/db"
)

// hostsDir is where host files are stored. This directory is created
// automatically if it doesn't exist.
var hostsDir = filepath.Join(os.TempDir(), "pssh-hosts")

// ParseHosts splits the provided string on newlines, commas and
// spaces, trimming whitespace and discarding empty entries.
func ParseHosts(in string) []string {
    var hosts []string
    // Replace common separators with newlines
    replaced := strings.NewReplacer(",", "\n", " ", "\n", "\r", "\n").Replace(in)
    scanner := bufio.NewScanner(strings.NewReader(replaced))
    for scanner.Scan() {
        h := strings.TrimSpace(scanner.Text())
        if h != "" {
            hosts = append(hosts, h)
        }
    }
    return hosts
}

// RunJob executes the command on the given hosts using parallel-ssh. It
// updates the database with results and job status.
func RunJob(id string, hosts []string, command string, timeout int) {
    // Create hosts directory
    if err := os.MkdirAll(hostsDir, 0o755); err != nil {
        db.UpdateJobStatus(id, "error", time.Now().UTC())
        return
    }
    // Create hosts file
    hostsFile := filepath.Join(hostsDir, id+".hosts")
    if err := os.WriteFile(hostsFile, []byte(strings.Join(hosts, "\n")), 0o644); err != nil {
        db.UpdateJobStatus(id, "error", time.Now().UTC())
        return
    }
    // Prepare context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
    defer cancel()
    // Build command
    args := []string{"-i", "-h", hostsFile, command}
    cmd := exec.CommandContext(ctx, "parallel-ssh", args...)
    start := time.Now()
    output, err := cmd.CombinedOutput()
    duration := time.Since(start)
    // Determine status and exit code
    status := "completed"
    exitCode := 0
    errMsg := ""
    if err != nil {
        // Determine exit code when possible
        status = "error"
        errMsg = err.Error()
        if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
            exitCode = exitErr.ExitCode()
        } else {
            exitCode = -1
        }
    }
    // For simplicity, store a single result entry with host "all"
    _ = db.InsertResult(id, "all", string(output), errMsg, exitCode, duration.Milliseconds())
    // Update job status
    _ = db.UpdateJobStatus(id, status, time.Now().UTC())
}
