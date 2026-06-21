package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/siddhantrao23/iota/internal/orchestrator"
)

type RunRequest struct {
	Type    string          `json:"type"`
	Args    json.RawMessage `json:"args"`
	Timeout int             `json:"timeout_ms,omitempty"`
}

type RunResponse struct {
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	ExitCode     int    `json:"exit_code"`
	DurationMs   int64  `json:"duration_ms"`
	CacheHit     bool   `json:"cache_hit"`
	InvocationID string `json:"invocation_id"`
}

var typeToRuntime = map[string]string{
	"code":       "", // resolved from args.language
	"shell":      "shell",
	"read_file":  "shell",
	"write_file": "shell",
	"list_dir":   "shell",
	"grep":       "shell",
	"http":       "shell",
}

func resolveRuntime(reqType string, args json.RawMessage) string {
	rt, ok := typeToRuntime[reqType]
	if !ok {
		return "shell"
	}
	if reqType == "code" {
		var ca struct {
			Language string `json:"language"`
		}
		json.Unmarshal(args, &ca)
		switch ca.Language {
		case "javascript", "js":
			return "javascript"
		case "go":
			return "go"
		default:
			return "python"
		}
	}
	return rt
}

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.StaticFile("/", "./static/index.html")

	router.POST("/run", func(c *gin.Context) {
		var req RunRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		runtime := resolveRuntime(req.Type, req.Args)

		if val, hit := orchestrator.GetCache(runtime, req.Args); hit {
			if val.Error != nil {
				c.JSON(http.StatusOK, gin.H{
					"error":  val.Error.Error(),
					"cached": true,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"output": val.Output,
					"cached": true,
				})
			}
			return
		}

		resultChan := make(chan orchestrator.Result)
		job := orchestrator.Job{
			Type:       req.Type,
			Args:       req.Args,
			ResultChan: resultChan,
		}

		queue, ok := orchestrator.JobQueues[runtime]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported runtime: " + runtime})
			return
		}

		select {
		case queue <- job:
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "system too busy"})
			return
		}

		select {
		case res := <-resultChan:
			orchestrator.PutCache(runtime, req.Args, res)
			if res.Error != nil {
				c.JSON(http.StatusOK, gin.H{
					"error":  res.Error.Error(),
					"cached": false,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"output": res.Output,
					"cached": false,
				})
			}
		}
	})

	return router
}
