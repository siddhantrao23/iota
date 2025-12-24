package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
	"github.com/siddhantrao23/iota/internal/orchestrator"
)

type ExecutionRequest struct {
	Code string `json:"code" binding:"required"`
}

func SetupRouter(cli *client.Client) *gin.Engine {
	router := gin.Default()

	router.POST("/run", func(c *gin.Context) {
		var req ExecutionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		start := time.Now()
		output, err := orchestrator.ExecuteCode(cli, req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"output": output,
			"time":   time.Since(start).Milliseconds(),
		})
	})

	return router
}
