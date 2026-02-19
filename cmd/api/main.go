package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-stakes-api/internal/di"
	"github.com/opinedajr/micro-stakes-api/internal/shared/middleware"
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

	bankrollRoutes := r.Group("/bankrolls")
	bankrollRoutes.Use(middleware.AuthMiddleware(container.Config().Keycloak, container.AuthService(), container.Logger()))
	{
		bankrollRoutes.POST("", container.BankrollHandler().CreateBankroll)
		bankrollRoutes.GET("", container.BankrollHandler().ListBankrolls)
		bankrollRoutes.GET("/:bankrollId", container.BankrollHandler().GetBankroll)
		bankrollRoutes.PUT("/:bankrollId", container.BankrollHandler().UpdateBankroll)
		bankrollRoutes.POST("/:bankrollId/reset", container.BankrollHandler().ResetBankroll)
	}

	log.Fatal(r.Run(":3003"))
}
