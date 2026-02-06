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

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: nil,
		})
		return
	}

	output, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: nil,
		})
		return
	}

	output, err := h.service.RefreshToken(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input LogoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: nil,
		})
		return
	}

	output, err := h.service.Logout(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, ErrorOutput{
			Error: "User already exists",
			Code:  "USER_EXISTS",
		})
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, ErrorOutput{
			Error: "Invalid credentials",
			Code:  "INVALID_CREDENTIALS",
		})
	case errors.Is(err, ErrTokenGenerationFailed):
		h.logger.Error("token generation failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "Failed to generate tokens",
			Code:  "TOKEN_GENERATION_FAILED",
		})
	case errors.Is(err, ErrValidationFailed):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	case errors.Is(err, ErrIdentityProviderError):
		h.logger.Error("identity provider error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "Authentication service unavailable",
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
