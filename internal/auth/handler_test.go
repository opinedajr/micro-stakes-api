package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	customValidator "github.com/opinedajr/micro-stakes-api/internal/shared/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RegisterOutput), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = customValidator.RegisterCustomValidators(v)
	}

	tests := []struct {
		name               string
		requestBody        interface{}
		mockServiceSetup   func(*MockAuthService)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - valid registration",
			requestBody: RegisterInput{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "SecureP@ss123",
			},
			mockServiceSetup: func(service *MockAuthService) {
				output := &RegisterOutput{
					ID:       1,
					Email:    "john.doe@example.com",
					FullName: "John Doe",
					Message:  "User registered successfully",
				}
				service.On("Register", mock.Anything, mock.MatchedBy(func(input RegisterInput) bool {
					return input.Email == "john.doe@example.com"
				})).Return(output, nil)
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response RegisterOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "john.doe@example.com", response.Email)
				assert.Equal(t, "John Doe", response.FullName)
				assert.NotEmpty(t, response.ID)
			},
		},
		{
			name: "error - invalid request body",
			requestBody: map[string]string{
				"invalid": "data",
			},
			mockServiceSetup: func(service *MockAuthService) {
			},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response ErrorOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response.Code)
			},
		},
		{
			name: "error - user already exists",
			requestBody: RegisterInput{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@example.com",
				Password:  "AnotherP@ss456",
			},
			mockServiceSetup: func(service *MockAuthService) {
				service.On("Register", mock.Anything, mock.MatchedBy(func(input RegisterInput) bool {
					return input.Email == "jane.smith@example.com"
				})).Return(nil, ErrUserAlreadyExists)
			},
			expectedStatusCode: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response ErrorOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "USER_EXISTS", response.Code)
				assert.Equal(t, "User already exists", response.Error)
			},
		},
		{
			name: "error - validation failed",
			requestBody: RegisterInput{
				FirstName: "Bob",
				LastName:  "Johnson",
				Email:     "bob.johnson@example.com",
				Password:  "ValidP@ss789",
			},
			mockServiceSetup: func(service *MockAuthService) {
				service.On("Register", mock.Anything, mock.MatchedBy(func(input RegisterInput) bool {
					return input.Email == "bob.johnson@example.com"
				})).Return(nil, WrapError(ErrValidationFailed, "invalid password format"))
			},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response ErrorOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response.Code)
			},
		},
		{
			name: "error - identity provider error",
			requestBody: RegisterInput{
				FirstName: "Alice",
				LastName:  "Brown",
				Email:     "alice.brown@example.com",
				Password:  "TestP@ss321",
			},
			mockServiceSetup: func(service *MockAuthService) {
				service.On("Register", mock.Anything, mock.MatchedBy(func(input RegisterInput) bool {
					return input.Email == "alice.brown@example.com"
				})).Return(nil, WrapError(ErrIdentityProviderError, "keycloak unavailable"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response ErrorOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "IDENTITY_PROVIDER_ERROR", response.Code)
				assert.Equal(t, "Failed to create user account", response.Error)
			},
		},
		{
			name: "error - database error",
			requestBody: RegisterInput{
				FirstName: "Charlie",
				LastName:  "Davis",
				Email:     "charlie.davis@example.com",
				Password:  "SecureP@ss999",
			},
			mockServiceSetup: func(service *MockAuthService) {
				service.On("Register", mock.Anything, mock.MatchedBy(func(input RegisterInput) bool {
					return input.Email == "charlie.davis@example.com"
				})).Return(nil, WrapError(ErrDatabaseError, "connection lost"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response ErrorOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "DATABASE_ERROR", response.Code)
				assert.Equal(t, "Database error occurred", response.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockServiceSetup(mockService)

			handler := NewAuthHandler(mockService, logger)

			router := gin.New()
			router.POST("/auth/register", handler.Register)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			tt.validateResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}
