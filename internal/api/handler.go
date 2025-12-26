package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/siddhantrao23/iota/internal/orchestrator"
)

type ExecutionRequest struct {
	Code string `json:"code" binding:"required"`
}

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/run", func(c *gin.Context) {
		var req ExecutionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resultChan := make(chan orchestrator.Result)
		job := orchestrator.Job{
			Code:       req.Code,
			ResultChan: resultChan,
		}

		select {
		case orchestrator.JobQueue <- job:
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "system too busy"})
			return
		}

		select {
		case res := <-resultChan:
			if res.Error != nil {
				print(res.Error.Error())
				c.JSON(http.StatusOK, gin.H{"error": res.Error.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"output": res.Output})
			}
		}
	})

	return router
}
