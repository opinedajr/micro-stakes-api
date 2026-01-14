package healthcheck

import (
	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	health := CheckHealth()
	c.JSON(200, health)
}
