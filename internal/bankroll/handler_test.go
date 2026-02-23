package bankroll

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

type MockBankrollServiceForHandler struct {
	mock.Mock
}

func (m *MockBankrollServiceForHandler) CreateBankroll(ctx context.Context, userID uint, input CreateBankrollInput) (*BankrollOutput, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BankrollOutput), args.Error(1)
}

func (m *MockBankrollServiceForHandler) UpdateBankroll(ctx context.Context, userID uint, bankrollID uint, input UpdateBankrollInput) (*BankrollOutput, error) {
	args := m.Called(ctx, userID, bankrollID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BankrollOutput), args.Error(1)
}

func (m *MockBankrollServiceForHandler) ListBankrolls(ctx context.Context, userID uint) ([]*BankrollOutput, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*BankrollOutput), args.Error(1)
}

func (m *MockBankrollServiceForHandler) GetBankroll(ctx context.Context, userID uint, bankrollID uint) (*BankrollOutput, error) {
	args := m.Called(ctx, userID, bankrollID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BankrollOutput), args.Error(1)
}

func (m *MockBankrollServiceForHandler) ResetBankroll(ctx context.Context, userID uint, bankrollID uint) (*BankrollOutput, error) {
	args := m.Called(ctx, userID, bankrollID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BankrollOutput), args.Error(1)
}

func TestCreateBankrollHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutput := &BankrollOutput{
			ID:                   1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		}

		mockService.On("CreateBankroll", mock.Anything, uint(1), mock.MatchedBy(func(input CreateBankrollInput) bool {
			return input.Name == "Main Bankroll" &&
				input.Currency == CurrencyBRL &&
				input.InitialBalance == 1000.00 &&
				input.StartDate == "2026-02-01" &&
				input.CommissionPercentage == 5.0
		})).Return(expectedOutput, nil).Once()

		requestBody := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateBankroll(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response BankrollOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "Main Bankroll", response.Name)
		assert.Equal(t, CurrencyBRL, response.Currency)
		assert.Equal(t, 1000.00, response.InitialBalance)
		assert.Equal(t, 1000.00, response.CurrentBalance)

		mockService.AssertExpectations(t)
	})

	t.Run("validation error - invalid JSON", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls", bytes.NewBuffer([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateBankroll(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid request body", response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Code)

		mockService.AssertNotCalled(t, "CreateBankroll")
	})

	t.Run("validation error - missing required field", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		requestBody := CreateBankrollInput{
			Name: "Main Bankroll",
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateBankroll(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "CreateBankroll")
	})

	t.Run("service error - duplicate name", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("CreateBankroll", mock.Anything, uint(1), mock.AnythingOfType("bankroll.CreateBankrollInput")).Return(nil, ErrBankrollNameExists).Once()

		requestBody := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateBankroll(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Bankroll name already exists", response.Error)
		assert.Equal(t, "BANKROLL_NAME_EXISTS", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("service error - validation failed", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("CreateBankroll", mock.Anything, uint(1), mock.AnythingOfType("bankroll.CreateBankrollInput")).Return(nil, ErrValidationFailed).Once()

		requestBody := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.CreateBankroll(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "VALIDATION_ERROR", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - missing userID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		requestBody := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateBankroll(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "CreateBankroll")
	})
}

func TestUpdateBankrollHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutput := &BankrollOutput{
			ID:                   1,
			Name:                 "Updated Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		}

		mockService.On("UpdateBankroll", mock.Anything, uint(1), uint(1), mock.MatchedBy(func(input UpdateBankrollInput) bool {
			return input.Name == "Updated Bankroll" &&
				input.Currency == CurrencyBRL &&
				input.StartDate == "2026-02-01" &&
				input.CommissionPercentage == 3.0
		})).Return(expectedOutput, nil).Once()

		requestBody := UpdateBankrollInput{
			Name:                 "Updated Bankroll",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response BankrollOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "Updated Bankroll", response.Name)
		assert.Equal(t, CurrencyBRL, response.Currency)
		assert.Equal(t, 3.0, response.CommissionPercentage)
		assert.Equal(t, 1000.00, response.InitialBalance)
		assert.Equal(t, 1000.00, response.CurrentBalance)

		mockService.AssertExpectations(t)
	})

	t.Run("validation error - invalid JSON", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid request body", response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Code)

		mockService.AssertNotCalled(t, "UpdateBankroll")
	})

	t.Run("validation error - missing required field", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		requestBody := UpdateBankrollInput{
			Name: "Updated Bankroll",
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertNotCalled(t, "UpdateBankroll")
	})

	t.Run("service error - not found", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("UpdateBankroll", mock.Anything, uint(1), uint(1), mock.AnythingOfType("bankroll.UpdateBankrollInput")).Return(nil, ErrBankrollNotFound).Once()

		requestBody := UpdateBankrollInput{
			Name:                 "Updated Bankroll",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Bankroll not found", response.Error)
		assert.Equal(t, "BANKROLL_NOT_FOUND", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("service error - duplicate name", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("UpdateBankroll", mock.Anything, uint(1), uint(1), mock.AnythingOfType("bankroll.UpdateBankrollInput")).Return(nil, ErrBankrollNameExists).Once()

		requestBody := UpdateBankrollInput{
			Name:                 "Updated Bankroll",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Bankroll name already exists", response.Error)
		assert.Equal(t, "BANKROLL_NAME_EXISTS", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("service error - validation failed", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("UpdateBankroll", mock.Anything, uint(1), uint(1), mock.AnythingOfType("bankroll.UpdateBankrollInput")).Return(nil, ErrValidationFailed).Once()

		requestBody := UpdateBankrollInput{
			Name:                 "Updated Bankroll",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "VALIDATION_ERROR", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - missing userID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		requestBody := UpdateBankrollInput{
			Name:                 "Updated Bankroll",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/bankrolls/1", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}

		handler.UpdateBankroll(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "UpdateBankroll")
	})
}

func TestListBankrollsHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutputs := []*BankrollOutput{
			{
				ID:                   1,
				Name:                 "Bankroll 1",
				Currency:             CurrencyBRL,
				InitialBalance:       1000.00,
				CurrentBalance:       1000.00,
				StartDate:            "2026-02-01",
				CommissionPercentage: 5.0,
				CreatedAt:            createdAt,
				UpdatedAt:            updatedAt,
			},
			{
				ID:                   2,
				Name:                 "Bankroll 2",
				Currency:             CurrencyUSD,
				InitialBalance:       500.00,
				CurrentBalance:       500.00,
				StartDate:            "2026-02-01",
				CommissionPercentage: 3.0,
				CreatedAt:            createdAt,
				UpdatedAt:            updatedAt,
			},
		}

		mockService.On("ListBankrolls", mock.Anything, uint(1)).Return(expectedOutputs, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/bankrolls", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.ListBankrolls(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*BankrollOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response, 2)
		assert.Equal(t, "Bankroll 1", response[0].Name)
		assert.Equal(t, "Bankroll 2", response[1].Name)

		mockService.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("ListBankrolls", mock.Anything, uint(1)).Return([]*BankrollOutput{}, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/bankrolls", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.ListBankrolls(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*BankrollOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Empty(t, response)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("ListBankrolls", mock.Anything, uint(1)).Return(nil, ErrDatabaseError).Once()

		req, err := http.NewRequest(http.MethodGet, "/bankrolls", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", "1")

		handler.ListBankrolls(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "DATABASE_ERROR", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - missing userID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodGet, "/bankrolls", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.ListBankrolls(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "ListBankrolls")
	})
}

func TestGetBankrollHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutput := &BankrollOutput{
			ID:                   1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		}

		mockService.On("GetBankroll", mock.Anything, uint(1), uint(1)).Return(expectedOutput, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/bankrolls/1", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.GetBankroll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response BankrollOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "Main Bankroll", response.Name)
		assert.Equal(t, CurrencyBRL, response.Currency)
		assert.Equal(t, 1000.00, response.InitialBalance)
		assert.Equal(t, 1000.00, response.CurrentBalance)

		mockService.AssertExpectations(t)
	})

	t.Run("bankroll not found", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("GetBankroll", mock.Anything, uint(1), uint(999)).Return(nil, ErrBankrollNotFound).Once()

		req, err := http.NewRequest(http.MethodGet, "/bankrolls/999", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "999"}}
		c.Set("userID", "1")

		handler.GetBankroll(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Bankroll not found", response.Error)
		assert.Equal(t, "BANKROLL_NOT_FOUND", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("GetBankroll", mock.Anything, uint(1), uint(1)).Return(nil, ErrDatabaseError).Once()

		req, err := http.NewRequest(http.MethodGet, "/bankrolls/1", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.GetBankroll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "DATABASE_ERROR", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - missing userID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodGet, "/bankrolls/1", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}

		handler.GetBankroll(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "GetBankroll")
	})

	t.Run("invalid bankroll ID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodGet, "/bankrolls/invalid", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "invalid"}}

		handler.GetBankroll(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "GetBankroll")
	})
}

func TestResetBankrollHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		createdAt := time.Now()
		updatedAt := time.Now()

		expectedOutput := &BankrollOutput{
			ID:                   1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       0.0,
			CurrentBalance:       0.0,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		}

		mockService.On("ResetBankroll", mock.Anything, uint(1), uint(1)).Return(expectedOutput, nil).Once()

		req, err := http.NewRequest(http.MethodPost, "/bankrolls/1/reset", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.ResetBankroll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response BankrollOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "Main Bankroll", response.Name)
		assert.Equal(t, 0.0, response.InitialBalance)
		assert.Equal(t, 0.0, response.CurrentBalance)

		mockService.AssertExpectations(t)
	})

	t.Run("bankroll not found", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("ResetBankroll", mock.Anything, uint(1), uint(999)).Return(nil, ErrBankrollNotFound).Once()

		req, err := http.NewRequest(http.MethodPost, "/bankrolls/999/reset", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "999"}}
		c.Set("userID", "1")

		handler.ResetBankroll(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Bankroll not found", response.Error)
		assert.Equal(t, "BANKROLL_NOT_FOUND", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		mockService.On("ResetBankroll", mock.Anything, uint(1), uint(1)).Return(nil, ErrDatabaseError).Once()

		req, err := http.NewRequest(http.MethodPost, "/bankrolls/1/reset", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}
		c.Set("userID", "1")

		handler.ResetBankroll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "DATABASE_ERROR", response.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - missing userID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls/1/reset", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "1"}}

		handler.ResetBankroll(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "ResetBankroll")
	})

	t.Run("invalid bankroll ID", func(t *testing.T) {
		mockService := new(MockBankrollServiceForHandler)
		logger := slog.Default()
		handler := NewBankrollHandler(mockService, logger)

		req, err := http.NewRequest(http.MethodPost, "/bankrolls/invalid/reset", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{gin.Param{Key: "bankrollId", Value: "invalid"}}
		c.Set("userID", "1")

		handler.ResetBankroll(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorOutput
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Unauthorized access to bankroll", response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Code)

		mockService.AssertNotCalled(t, "ResetBankroll")
	})
}
