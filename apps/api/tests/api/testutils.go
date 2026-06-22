package apitests

import (
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/models"
)

// =====================================================================
// Test Configuration & Fixtures
// =====================================================================

// TestConfig holds test configuration values
type TestConfig struct {
	JWTSecret       []byte
	JWTRefresh      []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// DefaultTestConfig returns a standard test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		JWTSecret:       []byte("test-secret-key-minimum-32-bytes-long!!"),
		JWTRefresh:      []byte("test-refresh-secret-minimum-32-bytes!!"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
}

// NewJWTManagerForTest creates a JWT manager with test configuration
func NewJWTManagerForTest(cfg *TestConfig) *auth.JWTManager {
	return auth.NewJWTManager(cfg.JWTSecret, cfg.JWTRefresh, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
}

// =====================================================================
// User Fixtures
// ========================================================================

// TestUser represents a test user fixture
type TestUser struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     models.Role
	MFA      bool
}

// AdminTestUser returns an admin test user fixture
func AdminTestUser() *TestUser {
	return &TestUser{
		ID:       uuid.New(),
		Email:    "admin@example.com",
		Password: "admin-password-12345678",
		Role:     models.RoleAdmin,
		MFA:      true,
	}
}

// OperatorTestUser returns an operator test user fixture
func OperatorTestUser() *TestUser {
	return &TestUser{
		ID:       uuid.New(),
		Email:    "operator@example.com",
		Password: "operator-password-123456",
		Role:     models.RoleOperator,
		MFA:      false,
	}
}

// ViewerTestUser returns a viewer test user fixture
func ViewerTestUser() *TestUser {
	return &TestUser{
		ID:       uuid.New(),
		Email:    "viewer@example.com",
		Password: "viewer-password-12345678",
		Role:     models.RoleViewer,
		MFA:      false,
	}
}

// GenerateTokenPair generates access + refresh tokens for a test user
func (u *TestUser) GenerateTokenPair(cfg *TestConfig) (string, string) {
	mgr := NewJWTManagerForTest(cfg)
	pair, _ := mgr.GenerateTokenPair(u.ID, u.Email, string(u.Role), u.MFA, u.MFA, "test-device")
	return pair.AccessToken, pair.RefreshToken
}

// =====================================================================
// Host Fixtures
// ========================================================================

// TestHost represents a test host fixture
type TestHost struct {
	ID       uuid.UUID
	OwnerID  uuid.UUID
	Name     string
	Address  string
	Protocol string
	Port     int
	Tags     []string
	Group    string
}

// ValidSSHHost returns a valid SSH host fixture
func ValidSSHHost(ownerID uuid.UUID) *TestHost {
	return &TestHost{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		Name:     "prod-ssh-server",
		Address:  "10.0.1.100",
		Protocol: "ssh",
		Port:     22,
		Tags:     []string{"production", "ssh"},
		Group:    "/prod",
	}
}

// ValidRDPHost returns a valid RDP host fixture
func ValidRDPHost(ownerID uuid.UUID) *TestHost {
	return &TestHost{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		Name:     "windows-prod",
		Address:  "10.0.2.50",
		Protocol: "rdp",
		Port:     3389,
		Tags:     []string{"production", "windows"},
		Group:    "/prod/windows",
	}
}

// ValidVNCHost returns a valid VNC host fixture
func ValidVNCHost(ownerID uuid.UUID) *TestHost {
	return &TestHost{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		Name:     "vnc-server",
		Address:  "10.0.3.25",
		Protocol: "vnc",
		Port:     5900,
		Tags:     []string{"dev"},
		Group:    "/dev",
	}
}

// =====================================================================
// Credential Fixtures
// ========================================================================

// TestCredential represents a test credential fixture
type TestCredential struct {
	ID     uuid.UUID
	OwnerID uuid.UUID
	Name   string
	Type   models.CredentialType
	Data   string
}

// ValidPasswordCredential returns a password credential fixture
func ValidPasswordCredential(ownerID uuid.UUID) *TestCredential {
	return &TestCredential{
		ID:     uuid.New(),
		OwnerID: ownerID,
		Name:   "prod-db-password",
		Type:   models.CredentialTypePassword,
		Data:   "super-secret-db-password-123",
	}
}

// ValidSSHKeyCredential returns an SSH key credential fixture
func ValidSSHKeyCredential(ownerID uuid.UUID) *TestCredential {
	return &TestCredential{
		ID:     uuid.New(),
		OwnerID: ownerID,
		Name:   "prod-ssh-key",
		Type:   models.CredentialTypeSSHKey,
		Data:   "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA0Z3VS5K...\n-----END RSA PRIVATE KEY-----",
	}
}

// ValidAPITokenCredential returns an API token credential fixture
func ValidAPITokenCredential(ownerID uuid.UUID) *TestCredential {
	return &TestCredential{
		ID:     uuid.New(),
		OwnerID: ownerID,
		Name:   "aws-api-token",
		Type:   models.CredentialTypeAPIToken,
		Data:   "AKIAIOSFODNN7EXAMPLE",
	}
}

// =====================================================================
// HTTP Request Helpers
// ========================================================================

// TestRequest holds components for building an HTTP test request
type TestRequest struct {
	Method string
	Path   string
	Body   interface{}
	Token  string
	Header map[string]string
}

// NewTestRequest creates a new test request
func NewTestRequest(method, path string) *TestRequest {
	return &TestRequest{
		Method: method,
		Path:   path,
		Header: make(map[string]string),
	}
}

// WithBody sets the request body
func (r *TestRequest) WithBody(body interface{}) *TestRequest {
	r.Body = body
	return r
}

// WithToken sets the authorization token
func (r *TestRequest) WithToken(token string) *TestRequest {
	r.Token = token
	return r
}

// WithHeader adds a header
func (r *TestRequest) WithHeader(key, value string) *TestRequest {
	r.Header[key] = value
	return r
}

// Execute sends the request and returns the response
func (r *TestRequest) Execute(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var body io.Reader
	if r.Body != nil {
		if s, ok := r.Body.(string); ok {
			body = strings.NewReader(s)
		} else {
			b, _ := json.Marshal(r.Body)
			body = strings.NewReader(string(b))
		}
	}
	req, _ := nethttp.NewRequest(r.Method, r.Path, body)
	if r.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	for k, v := range r.Header {
		req.Header.Set(k, v)
	}

	gin.SetMode(gin.TestMode)
 Router := gin.New()
	Router.Handle(r.Method, r.Path, handler)
	Router.ServeHTTP(w, req)
	return w
}

// =====================================================================
// Response Assertion Helpers
// ========================================================================

// AssertResponse asserts common response properties
func AssertResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, methodAndPath ...string) map[string]interface{} {
	if len(methodAndPath) == 2 { assert.Equal(t, expectedStatus, w.Code, "status code mismatch for %s %s", methodAndPath[0], methodAndPath[1]) } else { assert.Equal(t, expectedStatus, w.Code) }
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")
	return resp
}

// AssertErrorResponse asserts an error response structure
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMsgPart string) {
	resp := AssertResponse(t, w, expectedStatus)
	errMsg, ok := resp["error"].(string)
	require.True(t, ok, "response should have an 'error' string field")
	assert.Contains(t, errMsg, expectedMsgPart, "error message should contain '%s'", expectedMsgPart)
}

// AssertSuccessResponse asserts a success response structure
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	resp := AssertResponse(t, w, expectedStatus)
	assert.NotEqual(t, 0, len(resp), "success response should not be empty")
	return resp
}

// =====================================================================
// Validation Helpers
// ========================================================================

// ValidateUUID validates that a string is a valid UUID
func ValidateUUID(t *testing.T, s string) uuid.UUID {
	id, err := uuid.Parse(s)
	require.NoError(t, err, "should be a valid UUID: %s", s)
	return id
}

// ValidateTimestamp validates that a string is a parseable ISO8601 timestamp
func ValidateTimestamp(t *testing.T, s string) time.Time {
	ts, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err, "should be a valid timestamp: %s", s)
	return ts
}

// =====================================================================
// Random Data Generators
// ========================================================================

// RandomHex generates a random hex string of given byte length
func RandomHex(n int) string {
	b := make([]byte, n)
	crand.Read(b)
	return hex.EncodeToString(b)
}

// RandomEmail generates a random email address
func RandomEmail() string {
	return fmt.Sprintf("test+%s@example.com", RandomHex(8))
}

// RandomIP generates a random RFC1918 IP address
func RandomIP() string {
	return fmt.Sprintf("10.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

// =====================================================================
// Mock Data Store
// ========================================================================

// InMemoryStore provides a simple in-memory data store for tests
type InMemoryStore struct {
	Users      map[uuid.UUID]*TestUser
	Hosts      map[uuid.UUID]*TestHost
	Creds      map[uuid.UUID]*TestCredential
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		Users: make(map[uuid.UUID]*TestUser),
		Hosts: make(map[uuid.UUID]*TestHost),
		Creds: make(map[uuid.UUID]*TestCredential),
	}
}

// AddUser adds a user to the store
func (s *InMemoryStore) AddUser(u *TestUser) {
	s.Users[u.ID] = u
}

// AddHost adds a host to the store
func (s *InMemoryStore) AddHost(h *TestHost) {
	s.Hosts[h.ID] = h
}

// AddCredential adds a credential to the store
func (s *InMemoryStore) AddCredential(c *TestCredential) {
	s.Creds[c.ID] = c
}

// =====================================================================
// Authenticated Endpoint Test Helper
// ========================================================================

// AuthenticatedRouter creates a Gin router pre-configured with auth middleware simulation
func AuthenticatedRouter(user *TestUser, handler gin.HandlerFunc) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Set("email", user.Email)
		c.Set("role", string(user.Role))
		handler(c)
	})
	return r, httptest.NewRecorder()
}

// =====================================================================
// Common Test Scenarios
// ========================================================================

// TestUnauthenticatedAccess tests that endpoints reject unauthenticated requests
// setupRouter creates a minimal test router.
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUnauthenticatedAccess(t *testing.T, method, path string, handler gin.HandlerFunc) {
	r := setupRouter()
	r.Handle(method, path, handler)

	w := httptest.NewRecorder()
	req, _ := nethttp.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, nethttp.StatusUnauthorized, w.Code, "unauthenticated %s %s should return 401", method, path)
}

// TestNotFoundAccess tests that endpoints return 404 for non-existent resources
func TestNotFoundAccess(t *testing.T, method, path string, handler gin.HandlerFunc, token string) {
	r := setupRouter()
	r.Handle(method, path, handler)

	w := httptest.NewRecorder()
	req, _ := nethttp.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)

	assert.Equal(t, nethttp.StatusNotFound, w.Code, "%s %s should return 404 for non-existent resource", method, path)
}

// TestInvalidJSON tests that endpoints reject invalid JSON bodies
func TestInvalidJSON(t *testing.T, method, path string, handler gin.HandlerFunc, token string) {
	r := setupRouter()
	r.Handle(method, path, handler)

	w := httptest.NewRecorder()
	req, _ := nethttp.NewRequest(method, path, strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)

	assert.Equal(t, nethttp.StatusBadRequest, w.Code, "%s %s with invalid JSON should return 400", method, path)
}

// =====================================================================
// Integration Test Suite Template
// ========================================================================

// APITestSuite is a template for running comprehensive API tests
type APITestSuite struct {
	Config   *TestConfig
	Store    *InMemoryStore
	Admin    *TestUser
	Operator *TestUser
	Viewer   *TestUser
}

// NewAPITestSuite creates a new test suite with fixtures
func NewAPITestSuite() *APITestSuite {
	cfg := DefaultTestConfig()
	store := NewInMemoryStore()

	admin := AdminTestUser()
	operator := OperatorTestUser()
	viewer := ViewerTestUser()

	store.AddUser(admin)
	store.AddUser(operator)
	store.AddUser(viewer)

	return &APITestSuite{
		Config:   cfg,
		Store:    store,
		Admin:    admin,
		Operator: operator,
		Viewer:   viewer,
	}
}

// AdminToken returns an admin access token
func (s *APITestSuite) AdminToken() string {
	token, _ := s.Admin.GenerateTokenPair(s.Config)
	return token
}

// OperatorToken returns an operator access token
func (s *APITestSuite) OperatorToken() string {
	token, _ := s.Operator.GenerateTokenPair(s.Config)
	return token
}

// ViewerToken returns a viewer access token
func (s *APITestSuite) ViewerToken() string {
	token, _ := s.Viewer.GenerateTokenPair(s.Config)
	return token
}

// =====================================================================
// Table-Driven Test Helper
// ========================================================================

// RunTestCases runs a table of test cases
func RunTestCases(t *testing.T, name string, cases []struct {
		Title    string
		Input    interface{}
		Expected int
		Setup    func() *httptest.ResponseRecorder
}) {
	for _, tc := range cases {
		t.Run(name+"/"+tc.Title, func(t *testing.T) {
			w := tc.Setup()
			assert.Equal(t, tc.Expected, w.Code)
		})
	}
}

// =====================================================================
// TestMain Configuration
// ========================================================================

// PreCheck verifies test dependencies are available
func PreCheck(t *testing.T) {
	// This is a placeholder for any pre-test checks
	// e.g., verifying required env vars, external services, etc.
	t.Log("Test pre-checks passed")
}