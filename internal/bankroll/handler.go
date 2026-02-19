package bankroll

import (
	"errors"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type BankrollHandler struct {
	service BankrollService
	logger  *slog.Logger
}

func NewBankrollHandler(service BankrollService, logger *slog.Logger) *BankrollHandler {
	return &BankrollHandler{
		service: service,
		logger:  logger,
	}
}

func (h *BankrollHandler) CreateBankroll(c *gin.Context) {
	var input CreateBankrollInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: nil,
		})
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output, err := h.service.CreateBankroll(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (h *BankrollHandler) ListBankrolls(c *gin.Context) {
	panic("not implemented")
}

func (h *BankrollHandler) GetBankroll(c *gin.Context) {
	panic("not implemented")
}

func (h *BankrollHandler) UpdateBankroll(c *gin.Context) {
	panic("not implemented")
}

func (h *BankrollHandler) ResetBankroll(c *gin.Context) {
	panic("not implemented")
}

func (h *BankrollHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrBankrollNotFound):
		c.JSON(http.StatusNotFound, ErrorOutput{
			Error: "Bankroll not found",
			Code:  "BANKROLL_NOT_FOUND",
		})
	case errors.Is(err, ErrBankrollNameExists):
		c.JSON(http.StatusConflict, ErrorOutput{
			Error: "Bankroll name already exists",
			Code:  "BANKROLL_NAME_EXISTS",
		})
	case errors.Is(err, ErrValidationFailed):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: err.Error(),
			Code:  "VALIDATION_ERROR",
		})
	case errors.Is(err, ErrDatabaseError):
		h.logger.Error("database error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "Database error occurred",
			Code:  "DATABASE_ERROR",
		})
	case errors.Is(err, ErrUnauthorized):
		c.JSON(http.StatusForbidden, ErrorOutput{
			Error: "Unauthorized access to bankroll",
			Code:  "UNAUTHORIZED",
		})
	case errors.Is(err, ErrInvalidCurrency):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "Invalid currency",
			Code:  "INVALID_CURRENCY",
		})
	case errors.Is(err, ErrNegativeBalance):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "Balance cannot be negative",
			Code:  "NEGATIVE_BALANCE",
		})
	case errors.Is(err, ErrInvalidCommission):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "Commission percentage must be between 0 and 100",
			Code:  "INVALID_COMMISSION",
		})
	case errors.Is(err, ErrCannotModifyBalance):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "Cannot modify initial or current balance on update",
			Code:  "CANNOT_MODIFY_BALANCE",
		})
	default:
		h.logger.Error("unexpected error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "An unexpected error occurred",
			Code:  "INTERNAL_ERROR",
		})
	}
}

func (h *BankrollHandler) getUserID(c *gin.Context) (uint, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return 0, ErrUnauthorized
	}

	userID, ok := userIDStr.(string)
	if !ok {
		return 0, ErrUnauthorized
	}

	parsedID, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, ErrUnauthorized
	}

	return uint(parsedID), nil
}
