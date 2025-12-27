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

	router.StaticFile("/", "./static/index.html")

	router.POST("/run", func(c *gin.Context) {
		var req ExecutionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if val, hit := orchestrator.GetCache(req.Code); hit {
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
			orchestrator.PutCache(req.Code, res)
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
