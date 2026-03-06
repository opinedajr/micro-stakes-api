package strategy

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/gin-gonic/gin"
)

var errInvalidID = errors.New("invalid id")

type StrategyHandler struct {
	service StrategyService
	logger  *slog.Logger
}

func NewStrategyHandler(service StrategyService, logger *slog.Logger) *StrategyHandler {
	return &StrategyHandler{
		service: service,
		logger:  logger,
	}
}

func (h *StrategyHandler) CreateStrategy(c *gin.Context) {
	defer func() { _ = c.Request.Body.Close() }()

	var input CreateStrategyInput
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

	output, err := h.service.CreateStrategy(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (h *StrategyHandler) ListStrategies(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	output, err := h.service.ListStrategies(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *StrategyHandler) GetStrategy(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	strategyIDStr := c.Param("strategyId")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.GetStrategy(c.Request.Context(), userID, uint(strategyID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *StrategyHandler) UpdateStrategy(c *gin.Context) {
	defer func() { _ = c.Request.Body.Close() }()

	var input UpdateStrategyInput
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

	strategyIDStr := c.Param("strategyId")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.UpdateStrategy(c.Request.Context(), userID, uint(strategyID), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *StrategyHandler) UpdateStrategyStatus(c *gin.Context) {
	defer func() { _ = c.Request.Body.Close() }()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "Invalid request body",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	var rawBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: nil,
		})
		return
	}

	if _, hasActive := rawBody["active"]; !hasActive {
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "active field is required",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	var input UpdateStrategyStatusInput
	if err := json.Unmarshal(bodyBytes, &input); err != nil {
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

	strategyIDStr := c.Param("strategyId")
	strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.UpdateStrategyStatus(c.Request.Context(), userID, uint(strategyID), input.Active)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *StrategyHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errInvalidID):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "invalid id",
			Code:  "INVALID_ID",
		})
	case errors.Is(err, ErrStrategyNotFound):
		c.JSON(http.StatusNotFound, ErrorOutput{
			Error: "strategy not found",
			Code:  "STRATEGY_NOT_FOUND",
		})
	case errors.Is(err, ErrValidationFailed):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: err.Error(),
			Code:  "VALIDATION_FAILED",
		})
	case errors.Is(err, ErrDatabaseError):
		h.logger.Error("database error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "database error",
			Code:  "DATABASE_ERROR",
		})
	case errors.Is(err, ErrInvalidStrategyType):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "invalid strategy type",
			Code:  "INVALID_STRATEGY_TYPE",
		})
	case errors.Is(err, ErrInvalidDefaultStake):
		c.JSON(http.StatusBadRequest, ErrorOutput{
			Error: "invalid default stake",
			Code:  "INVALID_DEFAULT_STAKE",
		})
	default:
		h.logger.Error("unexpected error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorOutput{
			Error: "An unexpected error occurred",
			Code:  "INTERNAL_ERROR",
		})
	}
}

func (h *StrategyHandler) getUserID(c *gin.Context) (uint, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return 0, errInvalidID
	}

	userID, ok := userIDStr.(string)
	if !ok {
		return 0, errInvalidID
	}

	parsedID, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, errInvalidID
	}

	return uint(parsedID), nil
}
