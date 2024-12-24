package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"

	"github.com/zhshih/ratelimiter/internal/distributed"
	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type APIHandler struct {
	RateLimiter *ratelimiter.RateLimiter
	RaftNode    *raft.Raft
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

	cmd := distributed.RateLimitCommand{
		Action:   distributed.Increment,
		ClientID: clientID,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": false, "error": fmt.Sprintf("%s", err)})
		return
	}

	applyFuture := h.RaftNode.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": false, "error": fmt.Sprintf("%s", err)})
		return
	}

	_, ok := applyFuture.Response().(*distributed.ApplyResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"result": false, "error": "Error response is not matched"})
		return
	}

	// if !h.RateLimiter.AllowRequest(clientID) {
	// 	c.JSON(http.StatusTooManyRequests, gin.H{"result": false, "error": "Quota exceeded."})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"result": true})
}

func (h *APIHandler) ResetQuotaHandler(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing client_id."})
		return
	}

	cmd := distributed.RateLimitCommand{
		Action:    distributed.Reset,
		ClientID:  clientID,
		ResetTime: time.Now().Unix(),
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": false, "error": fmt.Sprintf("%s", err)})
		return
	}

	applyFuture := h.RaftNode.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": false, "error": fmt.Sprintf("%s", err)})
		return
	}

	_, ok := applyFuture.Response().(*distributed.ApplyResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"result": false, "error": "Error response is not matched"})
		return
	}
	// h.RateLimiter.ResetQuota(clientID)
	c.JSON(http.StatusOK, gin.H{"result": true})
}
