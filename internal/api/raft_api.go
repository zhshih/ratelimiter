package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

type RaftHandler struct {
	RaftNode *raft.Raft
}

type JoinRequest struct {
	NodeID   string `json:"node_id"`
	RaftAddr string `json:"raft_address"`
}

func New(raftNode *raft.Raft) *RaftHandler {
	return &RaftHandler{
		RaftNode: raftNode,
	}
}

func (h *RaftHandler) JoinRaftHandler(c *gin.Context) {
	var req JoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Invalid request payload."})
		return
	}
	if req.NodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing node_id."})
		return
	}
	if req.RaftAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing raft_address."})
		return
	}
	if h.RaftNode.State() != raft.Leader {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"result": false, "error": "Not the leader."})
		return
	}
	configFuture := h.RaftNode.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			gin.H{"result": false, "error": fmt.Sprintf("Failed to get raft configuration: %s", err.Error())})
		return
	}
	// This must be run on the leader or it will fail.
	f := h.RaftNode.AddVoter(raft.ServerID(req.NodeID), raft.ServerAddress(req.RaftAddr), 0, 0)
	if f.Error() != nil {
		c.JSON(http.StatusUnprocessableEntity,
			gin.H{"result": false, "error": fmt.Sprintf("Error of adding voter: %s", f.Error().Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"message": fmt.Sprintf("Node %s at %s joined successfully", req.NodeID, req.RaftAddr),
		"stats":   h.RaftNode.Stats()})
}

func (h *RaftHandler) RemoveRaftHandler(c *gin.Context) {
	var req JoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Invalid request payload."})
		return
	}
	if req.NodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": false, "error": "Missing node_id."})
		return
	}
	if h.RaftNode.State() != raft.Leader {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"result": false, "error": "Not the leader"})
		return
	}
	configFuture := h.RaftNode.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			gin.H{"result": false, "error": fmt.Sprintf("Failed to get raft configuration: %s", err.Error())})
		return
	}
	future := h.RaftNode.RemoveServer(raft.ServerID(req.NodeID), 0, 0)
	if err := future.Error(); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			gin.H{"result": false, "error": fmt.Sprintf("Error of removing existing node %s: %s", req.NodeID, err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"message": fmt.Sprintf("node %s removed successfully", req.NodeID),
		"stats":   h.RaftNode.Stats()})
}

func (h *RaftHandler) StatsRaftHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"stats":  h.RaftNode.Stats()})
}
