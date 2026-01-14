package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-stakes-api/internal/healthcheck"
)

func main() {
	r := gin.Default()

	r.GET("/health", healthcheck.HealthHandler)

	log.Fatal(r.Run(":3003"))
}
