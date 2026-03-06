package strategy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStrategyServiceForHandler struct {
	mock.Mock
}

func (m *MockStrategyServiceForHandler) CreateStrategy(ctx context.Context, userID uint, input CreateStrategyInput) (*StrategyOutput, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StrategyOutput), args.Error(1)
}

func (m *MockStrategyServiceForHandler) ListStrategies(ctx context.Context, userID uint, page int, pageSize int) (*StrategyListOutput, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StrategyListOutput), args.Error(1)
}

func (m *MockStrategyServiceForHandler) GetStrategy(ctx context.Context, userID uint, strategyID uint) (*StrategyOutput, error) {
	args := m.Called(ctx, userID, strategyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StrategyOutput), args.Error(1)
}

func (m *MockStrategyServiceForHandler) UpdateStrategy(ctx context.Context, userID uint, strategyID uint, input UpdateStrategyInput) (*StrategyOutput, error) {
	args := m.Called(ctx, userID, strategyID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StrategyOutput), args.Error(1)
}

func (m *MockStrategyServiceForHandler) UpdateStrategyStatus(ctx context.Context, userID uint, strategyID uint, active bool) (*StrategyOutput, error) {
	args := m.Called(ctx, userID, strategyID, active)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StrategyOutput), args.Error(1)
}

func TestCreateStrategyHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()
		desc := "Back high-value selections at odds > 2.0"

		expectedOutput := &StrategyOutput{
			ID:           1,
			UserID:       1,
			Name:         "Value Back Strategy",
			Description:  &desc,
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		mockService.On("CreateStrategy", mock.Anything, uint(1), mock.MatchedBy(func(input CreateStrategyInput) bool {
			return input.Name == "Value Back Strategy" &&
				input.Description != nil &&
				*input.Description == "Back high-value selections at odds > 2.0" &&
				input.DefaultStake == 10.00 &&
				input.Type == StrategyTypeBack &&
				input.Active != nil &&
				*input.Active == true
		})).Return(expectedOutput, nil).Once()

		requestBody := CreateStrategyInput{
			Name:         "Value Back Strategy",
			Description:  &desc,
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/strategies", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateStrategy(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response StrategyOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "Value Back Strategy", response.Name)
		assert.Equal(t, "Back high-value selections at odds > 2.0", *response.Description)
		assert.Equal(t, 10.00, response.DefaultStake)
		assert.Equal(t, StrategyTypeBack, response.Type)
		assert.True(t, response.Active)

		mockService.AssertExpectations(t)
	})

	t.Run("validation error - missing required field", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := map[string]interface{}{
			"name": "Test Strategy",
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/strategies", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid request body", response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Code)

		mockService.AssertNotCalled(t, "CreateStrategy")
	})

	t.Run("validation error - invalid type", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         "InvalidType",
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/strategies", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "CreateStrategy")
	})

	t.Run("service error - database error", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		mockService.On("CreateStrategy", mock.Anything, uint(1), mock.AnythingOfType("strategy.CreateStrategyInput")).Return(nil, ErrDatabaseError).Once()

		requestBody := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/strategies", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateStrategy(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "database error", response.Error)
		assert.Equal(t, "DATABASE_ERROR", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("service error - invalid strategy type", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		mockService.On("CreateStrategy", mock.Anything, uint(1), mock.AnythingOfType("strategy.CreateStrategyInput")).Return(nil, ErrInvalidStrategyType).Once()

		requestBody := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/strategies", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "invalid strategy type", response.Error)
		assert.Equal(t, "INVALID_STRATEGY_TYPE", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated - missing userID", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/strategies", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "invalid id", response.Error)
		assert.Equal(t, "INVALID_ID", response.Code)

		mockService.AssertNotCalled(t, "CreateStrategy")
	})
}

func TestListStrategiesHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()
		desc1 := "First strategy"
		desc2 := "Second strategy"

		expectedOutput := &StrategyListOutput{
			Data: []*StrategyOutput{
				{
					ID:           2,
					UserID:       1,
					Name:         "Strategy 2",
					Description:  &desc2,
					DefaultStake: 20.00,
					Type:         StrategyTypeLay,
					Active:       false,
					CreatedAt:    createdAt,
					UpdatedAt:    updatedAt,
				},
				{
					ID:           1,
					UserID:       1,
					Name:         "Strategy 1",
					Description:  &desc1,
					DefaultStake: 10.00,
					Type:         StrategyTypeBack,
					Active:       true,
					CreatedAt:    createdAt,
					UpdatedAt:    updatedAt,
				},
			},
			Total:      25,
			Page:       1,
			PageSize:   20,
			TotalPages: 2,
		}

		mockService.On("ListStrategies", mock.Anything, uint(1), 1, 20).Return(expectedOutput, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/strategies?page=1&page_size=20", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.ListStrategies(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StrategyListOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, int64(25), response.Total)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.PageSize)
		assert.Equal(t, 2, response.TotalPages)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, "Strategy 2", response.Data[0].Name)
		assert.Equal(t, "Strategy 1", response.Data[1].Name)

		mockService.AssertExpectations(t)
	})

	t.Run("page and page_size query params parsed correctly", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		expectedOutput := &StrategyListOutput{
			Data:       []*StrategyOutput{},
			Total:      0,
			Page:       2,
			PageSize:   10,
			TotalPages: 0,
		}

		mockService.On("ListStrategies", mock.Anything, uint(1), 2, 10).Return(expectedOutput, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/strategies?page=2&page_size=10", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.ListStrategies(c)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("out-of-bounds page returns 200 with empty data", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		expectedOutput := &StrategyListOutput{
			Data:       []*StrategyOutput{},
			Total:      5,
			Page:       9999,
			PageSize:   20,
			TotalPages: 1,
		}

		mockService.On("ListStrategies", mock.Anything, uint(1), 9999, 20).Return(expectedOutput, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/strategies?page=9999&page_size=20", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.ListStrategies(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StrategyListOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, int64(5), response.Total)
		assert.Equal(t, 1, response.TotalPages)
		assert.Len(t, response.Data, 0)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated - missing userID", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodGet, "/strategies", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.ListStrategies(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "invalid id", response.Error)
		assert.Equal(t, "INVALID_ID", response.Code)

		mockService.AssertNotCalled(t, "ListStrategies")
	})
}

func TestGetStrategyHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()
		desc := "Back high-value selections at odds > 2.0"

		expectedOutput := &StrategyOutput{
			ID:           123,
			UserID:       1,
			Name:         "Value Back Strategy",
			Description:  &desc,
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		mockService.On("GetStrategy", mock.Anything, uint(1), uint(123)).Return(expectedOutput, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/strategies/123", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}
		c.Set("userID", "1")

		handler.GetStrategy(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StrategyOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(123), response.ID)
		assert.Equal(t, "Value Back Strategy", response.Name)
		assert.Equal(t, "Back high-value selections at odds > 2.0", *response.Description)
		assert.Equal(t, 10.00, response.DefaultStake)
		assert.Equal(t, StrategyTypeBack, response.Type)
		assert.True(t, response.Active)

		mockService.AssertExpectations(t)
	})

	t.Run("not found returns 404 with STRATEGY_NOT_FOUND code", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		mockService.On("GetStrategy", mock.Anything, uint(1), uint(99999)).Return(nil, ErrStrategyNotFound).Once()

		req, err := http.NewRequest(http.MethodGet, "/strategies/99999", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "99999"}}
		c.Set("userID", "1")

		handler.GetStrategy(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "strategy not found", response.Error)
		assert.Equal(t, "STRATEGY_NOT_FOUND", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated - missing userID", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodGet, "/strategies/123", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}

		handler.GetStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "invalid id", response.Error)
		assert.Equal(t, "INVALID_ID", response.Code)

		mockService.AssertNotCalled(t, "GetStrategy")
	})
}

func TestUpdateStrategyHandler(t *testing.T) {
	t.Run("success returns 200 with updated body", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()
		desc := "Refined criteria for value bets"

		expectedOutput := &StrategyOutput{
			ID:           123,
			UserID:       1,
			Name:         "Updated Value Strategy",
			Description:  &desc,
			DefaultStake: 15.00,
			Type:         StrategyTypeBack,
			Active:       true,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		mockService.On("UpdateStrategy", mock.Anything, uint(1), uint(123), mock.MatchedBy(func(input UpdateStrategyInput) bool {
			return input.Name == "Updated Value Strategy" &&
				input.Description != nil &&
				*input.Description == "Refined criteria for value bets" &&
				input.DefaultStake == 15.00 &&
				input.Type == StrategyTypeBack &&
				input.Active != nil &&
				*input.Active == true
		})).Return(expectedOutput, nil).Once()

		requestBody := UpdateStrategyInput{
			Name:         "Updated Value Strategy",
			Description:  &desc,
			DefaultStake: 15.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/strategies/123", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}
		c.Set("userID", "1")

		handler.UpdateStrategy(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StrategyOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(123), response.ID)
		assert.Equal(t, "Updated Value Strategy", response.Name)
		assert.Equal(t, "Refined criteria for value bets", *response.Description)
		assert.Equal(t, 15.00, response.DefaultStake)
		assert.Equal(t, StrategyTypeBack, response.Type)
		assert.True(t, response.Active)

		mockService.AssertExpectations(t)
	})

	t.Run("missing required field returns 400", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := map[string]interface{}{
			"description": "Missing name and type",
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/strategies/123", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}
		c.Set("userID", "1")

		handler.UpdateStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "UpdateStrategy")
	})

	t.Run("strategy not found returns 404", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		mockService.On("UpdateStrategy", mock.Anything, uint(1), uint(99999), mock.AnythingOfType("strategy.UpdateStrategyInput")).Return(nil, ErrStrategyNotFound).Once()

		requestBody := UpdateStrategyInput{
			Name:         "Updated Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/strategies/99999", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "99999"}}
		c.Set("userID", "1")

		handler.UpdateStrategy(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "strategy not found", response.Error)
		assert.Equal(t, "STRATEGY_NOT_FOUND", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := UpdateStrategyInput{
			Name:         "Updated Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/strategies/123", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}

		handler.UpdateStrategy(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "UpdateStrategy")
	})
}

func TestUpdateStrategyStatusHandler(t *testing.T) {
	t.Run("success activate returns 200 with active true", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutput := &StrategyOutput{
			ID:           123,
			UserID:       1,
			Name:         "Value Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		mockService.On("UpdateStrategyStatus", mock.Anything, uint(1), uint(123), true).Return(expectedOutput, nil).Once()

		requestBody := UpdateStrategyStatusInput{
			Active: true,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPatch, "/strategies/123/status", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}
		c.Set("userID", "1")

		handler.UpdateStrategyStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StrategyOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(123), response.ID)
		assert.True(t, response.Active)

		mockService.AssertExpectations(t)
	})

	t.Run("success deactivate returns 200 with active false", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutput := &StrategyOutput{
			ID:           123,
			UserID:       1,
			Name:         "Value Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       false,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		mockService.On("UpdateStrategyStatus", mock.Anything, uint(1), uint(123), false).Return(expectedOutput, nil).Once()

		requestBody := UpdateStrategyStatusInput{
			Active: false,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPatch, "/strategies/123/status", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}
		c.Set("userID", "1")

		handler.UpdateStrategyStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StrategyOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(123), response.ID)
		assert.False(t, response.Active)

		mockService.AssertExpectations(t)
	})

	t.Run("missing active field returns 400", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := map[string]interface{}{}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPatch, "/strategies/123/status", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}
		c.Set("userID", "1")

		handler.UpdateStrategyStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "UpdateStrategyStatus")
	})

	t.Run("strategy not found returns 404", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		mockService.On("UpdateStrategyStatus", mock.Anything, uint(1), uint(99999), true).Return(nil, ErrStrategyNotFound).Once()

		requestBody := UpdateStrategyStatusInput{
			Active: true,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPatch, "/strategies/99999/status", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "99999"}}
		c.Set("userID", "1")

		handler.UpdateStrategyStatus(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "strategy not found", response.Error)
		assert.Equal(t, "STRATEGY_NOT_FOUND", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		mockService := new(MockStrategyServiceForHandler)
		logger := slog.Default()
		handler := NewStrategyHandler(mockService, logger)

		requestBody := UpdateStrategyStatusInput{
			Active: true,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPatch, "/strategies/123/status", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "strategyId", Value: "123"}}

		handler.UpdateStrategyStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "UpdateStrategyStatus")
	})
}
