package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
	logger  *slog.Logger
}

func NewAuthHandler(service AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: nil,
		})
		return
	}

	output, err := h.service.Register(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, ErrorOutput{
			Error: "User already exists",
			Code:  "USER_EXISTS",
		})
	case errors.Is(err, ErrValidationFailed):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	case errors.Is(err, ErrIdentityProviderError):
		h.logger.Error("identity provider error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "Failed to create user account",
			Code:  "IDENTITY_PROVIDER_ERROR",
		})
	case errors.Is(err, ErrDatabaseError):
		h.logger.Error("database error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "Database error occurred",
			Code:  "DATABASE_ERROR",
		})
	default:
		h.logger.Error("unexpected error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "An unexpected error occurred",
			Code:  "INTERNAL_ERROR",
		})
	}
}
