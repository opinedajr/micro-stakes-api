package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-stakes-api/internal/di"
)

func main() {
	container := di.NewContainer()
	r := gin.Default()

	r.GET("/health", container.HealthCheckHandler().Handle)

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", container.AuthHandler().Register)
		authRoutes.POST("/login", container.AuthHandler().Login)
		authRoutes.POST("/refresh", container.AuthHandler().RefreshToken)
		authRoutes.POST("/logout", container.AuthHandler().Logout)
	}

	log.Fatal(r.Run(":3003"))
}
