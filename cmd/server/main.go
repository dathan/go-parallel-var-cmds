package main

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "github.com/dathan/go-parallel-var-cmds/internal/db"
    "github.com/dathan/go-parallel-var-cmds/internal/executor"
)

type RunRequest struct {
    Hosts   string `json:"hosts"`
    Command string `json:"command"`
    Timeout int    `json:"timeout"`
}

func main() {
    router := gin.Default()
    database, err := db.NewDatabase("jobs.db")
    if err != nil {
        panic(err)
    }

    router.POST("/api/run", func(c *gin.Context) {
        var req RunRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        jobID := uuid.New().String()
        now := time.Now().UTC()
        job := &db.Job{
            ID:        jobID,
            Command:   req.Command,
            CreatedAt: now,
            Status:    "running",
        }
        if err := database.CreateJob(job); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        hosts := executor.ParseHosts(req.Hosts)
        if err := database.InsertHosts(jobID, hosts); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // run in background
        go func() {
            executor.RunJob(database, jobID, hosts, req.Command, req.Timeout)
        }()

        c.JSON(http.StatusOK, gin.H{"id": jobID})
    })

    router.GET("/api/jobs", func(c *gin.Context) {
        jobs, err := database.ListJobs()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, jobs)
    })

    router.GET("/api/jobs/:id", func(c *gin.Context) {
        id := c.Param("id")
        job, err := database.GetJob(id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, job)
    })

    router.Run(":8080")
}
