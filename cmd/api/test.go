package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mineral/data"
	"mineral/handlers"
	"mineral/routes"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	// Create a test router
	router := routes.SetupRoutes(nil, nil, nil, nil, nil, nil)

	// Create a request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// TestSignupEndpoint tests the signup endpoint
func TestSignupEndpoint(t *testing.T) {
	// Create mock user repository
	userRepo := &MockUserRepository{}

	// Create auth handler
	authHandler := handlers.NewAuthHandler(userRepo)

	// Create a test router
	router := routes.SetupRoutes(authHandler, nil, nil, nil, nil, nil)

	// Create signup request
	signupReq := handlers.SignupRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	jsonData, err := json.Marshal(signupReq)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request
	req, err := http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

// MockUserRepository is a mock implementation for testing
type MockUserRepository struct{}

func (m *MockUserRepository) GetAll() ([]*data.User, error) {
	return []*data.User{}, nil
}

func (m *MockUserRepository) GetByEmail(email string) (*data.User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) GetOne(id uint) (*data.User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) Insert(user *data.User) (uint, error) {
	return 1, nil
}

func (m *MockUserRepository) Update(user *data.User) error {
	return nil
}

func (m *MockUserRepository) Delete(user *data.User) error {
	return nil
}

func (m *MockUserRepository) DeleteByID(id uint) error {
	return nil
}

func (m *MockUserRepository) ResetPassword(userID uint, newPassword string) error {
	return nil
}

func (m *MockUserRepository) PasswordMatches(user *data.User, plainText string) (bool, error) {
	return true, nil
}

func (m *MockUserRepository) GenerateAndSaveOTP(email string) (string, error) {
	return "123456", nil
}

func (m *MockUserRepository) VerifyOTP(email, otp string) (bool, error) {
	return true, nil
}

func (m *MockUserRepository) ResetPasswordWithOTP(email, otp, newPassword string) error {
	return nil
}
