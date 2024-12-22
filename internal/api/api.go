package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type APIHandler struct {
	RateLimiter *ratelimiter.RateLimiter
}

func (h *APIHandler) CheckQuotaHandler(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing client_id."})
		return
	}

	remaining := h.RateLimiter.CheckQuota(clientID)
	c.JSON(http.StatusOK, gin.H{"result": true, "remaining_quota": remaining})
}

func (h *APIHandler) IncrementQuotaHandler(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing client_id."})
		return
	}

	if !h.RateLimiter.AllowRequest(clientID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"result": false, "error": "Quota exceeded."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": true})
}

func (h *APIHandler) ResetQuotaHandler(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing client_id."})
		return
	}

	h.RateLimiter.ResetQuota(clientID)
	c.JSON(http.StatusOK, gin.H{"result": true})
}
