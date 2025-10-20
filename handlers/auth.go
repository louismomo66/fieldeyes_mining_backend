package handlers

import (
	"encoding/json"
	"fmt"
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	UserRepo data.UserInterface
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo data.UserInterface) *AuthHandler {
	return &AuthHandler{
		UserRepo: userRepo,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Phone     string `json:"phone,omitempty"`
	Password  string `json:"password"`
	AdminCode string `json:"admin_code,omitempty"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate input
	if !utils.ValidateEmail(req.Email) {
		utils.WriteValidationError(w, "Invalid email format")
		return
	}
	if !utils.ValidatePassword(req.Password) {
		utils.WriteValidationError(w, "Password must be at least 6 characters")
		return
	}

	// Get user by email
	user, err := h.UserRepo.GetByEmail(req.Email)
	if err != nil {
		utils.WriteUnauthorizedError(w, "Invalid email or password")
		return
	}

	// Check password
	valid, err := h.UserRepo.PasswordMatches(user, req.Password)
	if err != nil || !valid {
		utils.WriteUnauthorizedError(w, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.Email, string(user.Role))
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to generate token")
		return
	}

	// Return success response with token
	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone,
			"role":  user.Role,
		},
	}

	utils.WriteSuccessResponse(w, "Login successful", response)
}

// Signup handles user registration
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	// Handle both JSON and FormData
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" {
		// Parse form data first
		if err := r.ParseForm(); err != nil {
			utils.WriteValidationError(w, "Invalid form data")
			return
		}
		// Handle FormData
		req.Email = r.FormValue("email")
		req.Name = r.FormValue("name")
		req.Password = r.FormValue("password")
		req.Phone = r.FormValue("phone")
	} else {
		// Handle JSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteValidationError(w, "Invalid request body")
			return
		}
	}

	// Validate input
	if !utils.ValidateEmail(req.Email) {
		utils.WriteValidationError(w, "Invalid email format")
		return
	}
	if !utils.ValidateRequired(req.Name) {
		utils.WriteValidationError(w, "Name is required")
		return
	}
	if !utils.ValidatePassword(req.Password) {
		utils.WriteValidationError(w, "Password must be at least 6 characters")
		return
	}
	if req.Phone != "" && !utils.ValidatePhone(req.Phone) {
		utils.WriteValidationError(w, "Invalid phone number format")
		return
	}

	// Check if user already exists
	existingUser, _ := h.UserRepo.GetByEmail(req.Email)
	if existingUser != nil {
		utils.WriteValidationError(w, "Email already registered")
		return
	}

	// Determine user role
	role := data.RoleStandard
	if req.AdminCode != "" {
		if req.AdminCode == "MINING2025ADMIN" {
			role = data.RoleAdmin
		} else {
			utils.WriteValidationError(w, "Invalid admin code")
			return
		}
	}

	// Create new user
	user := &data.User{
		Email: req.Email,
		Name:  req.Name,
		Phone: &req.Phone,
		Role:  role,
	}
	if req.Phone == "" {
		user.Phone = nil
	}

	user.Password = req.Password // Will be hashed in repository

	userID, err := h.UserRepo.Insert(user)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(fmt.Sprintf("%d", userID), user.Email, string(user.Role))
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to generate token")
		return
	}

	// Return success response
	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    userID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone,
			"role":  user.Role,
		},
	}

	utils.WriteSuccessResponse(w, "User created successfully", response)
}

// ForgotPassword handles forgot password requests
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate email
	if !utils.ValidateEmail(req.Email) {
		utils.WriteValidationError(w, "Invalid email format")
		return
	}

	// Check if user exists
	_, err := h.UserRepo.GetByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		utils.WriteSuccessResponse(w, "If the email exists, an OTP has been sent", nil)
		return
	}

	// Generate and save OTP
	otp, err := h.UserRepo.GenerateAndSaveOTP(req.Email)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to generate OTP")
		return
	}

	// In a real application, you would send the OTP via email/SMS
	// For now, we'll just log it (remove this in production)
	// TODO: Replace with proper logging
	_ = otp // Suppress unused variable warning

	utils.WriteSuccessResponse(w, "If the email exists, an OTP has been sent", nil)
}

// ResetPassword handles password reset with OTP
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate input
	if !utils.ValidateEmail(req.Email) {
		utils.WriteValidationError(w, "Invalid email format")
		return
	}
	if !utils.ValidateRequired(req.OTP) {
		utils.WriteValidationError(w, "OTP is required")
		return
	}
	if !utils.ValidatePassword(req.NewPassword) {
		utils.WriteValidationError(w, "Password must be at least 6 characters")
		return
	}

	// Reset password with OTP
	err := h.UserRepo.ResetPasswordWithOTP(req.Email, req.OTP, req.NewPassword)
	if err != nil {
		utils.WriteValidationError(w, "Invalid or expired OTP")
		return
	}

	utils.WriteSuccessResponse(w, "Password reset successfully", nil)
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	user, err := h.UserRepo.GetOne(userID)
	if err != nil {
		utils.WriteNotFoundError(w, "User not found")
		return
	}

	// Remove sensitive information
	response := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"phone": user.Phone,
		"role":  user.Role,
	}

	utils.WriteSuccessResponse(w, "Profile retrieved successfully", response)
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	var req struct {
		Name     string  `json:"name"`
		Phone    *string `json:"phone,omitempty"`
		Location *string `json:"location,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate input
	if !utils.ValidateRequired(req.Name) {
		utils.WriteValidationError(w, "Name is required")
		return
	}
	if req.Phone != nil && *req.Phone != "" && !utils.ValidatePhone(*req.Phone) {
		utils.WriteValidationError(w, "Invalid phone number format")
		return
	}

	// Get current user
	user, err := h.UserRepo.GetOne(userID)
	if err != nil {
		utils.WriteNotFoundError(w, "User not found")
		return
	}

	// Update user
	user.Name = req.Name
	user.Phone = req.Phone
	user.Location = req.Location

	err = h.UserRepo.Update(user)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update profile")
		return
	}

	// Return updated profile
	response := map[string]interface{}{
		"id":       user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"phone":    user.Phone,
		"location": user.Location,
		"role":     user.Role,
	}

	utils.WriteSuccessResponse(w, "Profile updated successfully", response)
}
