package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/zhshih/ratelimiter/internal/api"
	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	rl := ratelimiter.NewRateLimiter()
	apiHandler := &api.APIHandler{
		RateLimiter: rl,
	}
	router.GET("/rate/check", apiHandler.CheckQuotaHandler)
	router.POST("/rate/increment", apiHandler.IncrementQuotaHandler)
	router.POST("/rate/reset", apiHandler.ResetQuotaHandler)

	log.Println("Rate limiter running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
