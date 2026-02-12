package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return privateKey, &privateKey.PublicKey
}

func createTestToken(t *testing.T, privateKey *rsa.PrivateKey, claims jwt.MapClaims, expiration time.Duration) string {
	t.Helper()

	if expiration != 0 {
		claims["exp"] = time.Now().Add(expiration).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	return tokenString
}

func createMockJWKSHandler(t *testing.T, publicKey *rsa.PublicKey) http.HandlerFunc {
	t.Helper()

	nBytes := publicKey.N.Bytes()
	nBase64 := base64.RawURLEncoding.EncodeToString(nBytes)

	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	eBase64 := base64.RawURLEncoding.EncodeToString(eBytes)

	jwks := JWKS{
		Keys: []JWK{
			{
				Kid: "test-key-id",
				Kty: "RSA",
				Alg: "RS256",
				Use: "sig",
				N:   nBase64,
				E:   eBase64,
			},
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}
}

func TestAuthMiddleware(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	mockServer := httptest.NewServer(createMockJWKSHandler(t, publicKey))
	defer mockServer.Close()

	cfg := config.KeycloakConfig{
		URL:   mockServer.URL,
		Realm: "test-realm",
	}

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		authHeader         string
		prepareToken       func() string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "error - missing authorization header",
			authHeader:         "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Authorization header required",
				"code":  "MISSING_TOKEN",
			},
		},
		{
			name:               "error - invalid authorization format (no Bearer)",
			authHeader:         "InvalidFormat token123",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Invalid authorization format",
				"code":  "INVALID_TOKEN_FORMAT",
			},
		},
		{
			name:               "error - invalid authorization format (too many parts)",
			authHeader:         "Bearer part1 part2 part3",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Invalid authorization format",
				"code":  "INVALID_TOKEN_FORMAT",
			},
		},
		{
			name:               "error - invalid token signature",
			authHeader:         "Bearer invalid.token.here",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			},
		},
		{
			name: "error - expired token",
			prepareToken: func() string {
				return createTestToken(t, privateKey, jwt.MapClaims{
					"sub":   "user-123",
					"email": "user@example.com",
				}, -1*time.Hour)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			},
		},
		{
			name: "error - missing kid in token header",
			prepareToken: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"sub":   "user-123",
					"email": "user@example.com",
					"exp":   time.Now().Add(1 * time.Hour).Unix(),
				})
				token.Header["kid"] = ""
				tokenString, err := token.SignedString(privateKey)
				require.NoError(t, err)
				return tokenString
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			},
		},
		{
			name: "error - wrong signing method",
			prepareToken: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub":   "user-123",
					"email": "user@example.com",
					"exp":   time.Now().Add(1 * time.Hour).Unix(),
				})
				token.Header["kid"] = "test-key-id"
				token.Header["alg"] = "HS256"
				tokenString, err := token.SignedString([]byte("secret"))
				require.NoError(t, err)
				return tokenString
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			},
		},
		{
			name: "success - valid token",
			prepareToken: func() string {
				return createTestToken(t, privateKey, jwt.MapClaims{
					"sub":   "user-123",
					"email": "test@example.com",
				}, 1*time.Hour)
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader := tt.authHeader
			if tt.prepareToken != nil {
				authHeader = fmt.Sprintf("Bearer %s", tt.prepareToken())
			}

			router := gin.New()
			router.Use(AuthMiddleware(cfg))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedResponse != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse["error"], response["error"])
				assert.Equal(t, tt.expectedResponse["code"], response["code"])
			}

			if tt.expectedStatusCode == http.StatusOK {
				routes := router.Routes()
				assert.True(t, len(routes) > 0, "route should be registered")
			}
		})
	}
}

func TestAuthMiddleware_UserIDAndEmailInContext(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	mockServer := httptest.NewServer(createMockJWKSHandler(t, publicKey))
	defer mockServer.Close()

	cfg := config.KeycloakConfig{
		URL:   mockServer.URL,
		Realm: "test-realm",
	}

	token := createTestToken(t, privateKey, jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
	}, 1*time.Hour)

	router := gin.New()
	router.Use(AuthMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "user-123", userID)

		email, exists := c.Get("email")
		assert.True(t, exists)
		assert.Equal(t, "test@example.com", email)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFetchPublicKey(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)

	mockServer := httptest.NewServer(createMockJWKSHandler(t, publicKey))
	defer mockServer.Close()

	tests := []struct {
		name          string
		prepareToken  func() *jwt.Token
		prepareConfig func() config.KeycloakConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "success - valid key fetched",
			prepareToken: func() *jwt.Token {
				token := jwt.New(jwt.SigningMethodRS256)
				token.Header["kid"] = "test-key-id"
				return token
			},
			prepareConfig: func() config.KeycloakConfig {
				return config.KeycloakConfig{
					URL:   mockServer.URL,
					Realm: "test-realm",
				}
			},
			expectError: false,
		},
		{
			name: "error - invalid JWKS URL",
			prepareToken: func() *jwt.Token {
				token := jwt.New(jwt.SigningMethodRS256)
				token.Header["kid"] = "test-key-id"
				return token
			},
			prepareConfig: func() config.KeycloakConfig {
				return config.KeycloakConfig{
					URL:   "invalid-keycloak-that-does-not-exist-123456789:9999",
					Realm: "test-realm",
				}
			},
			expectError:   true,
			errorContains: "failed to fetch JWKS",
		},
		{
			name: "error - missing kid in token",
			prepareToken: func() *jwt.Token {
				token := jwt.New(jwt.SigningMethodRS256)
				return token
			},
			prepareConfig: func() config.KeycloakConfig {
				return config.KeycloakConfig{
					URL:   mockServer.URL,
					Realm: "test-realm",
				}
			},
			expectError:   true,
			errorContains: "kid not found",
		},
		{
			name: "error - kid not found in JWKS",
			prepareToken: func() *jwt.Token {
				token := jwt.New(jwt.SigningMethodRS256)
				token.Header["kid"] = "non-existent-key-id"
				return token
			},
			prepareConfig: func() config.KeycloakConfig {
				return config.KeycloakConfig{
					URL:   mockServer.URL,
					Realm: "test-realm",
				}
			},
			expectError:   true,
			errorContains: "unable to find key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCfg := tt.prepareConfig()
			token := tt.prepareToken()
			_, err := fetchPublicKey(testCfg, token)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseRSAPublicKey(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)

	nBytes := publicKey.N.Bytes()
	nBase64 := base64.RawURLEncoding.EncodeToString(nBytes)

	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	eBase64 := base64.RawURLEncoding.EncodeToString(eBytes)

	validJWK := JWK{
		Kid: "test-key",
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   nBase64,
		E:   eBase64,
	}

	tests := []struct {
		name        string
		jwk         JWK
		expectError bool
	}{
		{
			name:        "success - valid JWK",
			jwk:         validJWK,
			expectError: false,
		},
		{
			name: "error - invalid base64 for N",
			jwk: JWK{
				N: "invalid-base64!!!",
				E: eBase64,
			},
			expectError: true,
		},
		{
			name: "error - invalid base64 for E",
			jwk: JWK{
				N: nBase64,
				E: "invalid-base64!!!",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedKey, err := parseRSAPublicKey(tt.jwk)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, parsedKey)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parsedKey)
				assert.Equal(t, publicKey.N, parsedKey.N)
				assert.Equal(t, publicKey.E, parsedKey.E)
			}
		})
	}
}
