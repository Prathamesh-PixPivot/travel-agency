// internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"travel-agency/internal/auth"
	"travel-agency/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UserID   uint   `json:"userId"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TenantID uint   `json:"tenantId"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	AccessToken         string `json:"accessToken"`
	RefreshToken        string `json:"refreshToken"`
	ForcePasswordChange bool   `json:"forcePasswordChange"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}
type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
}

type RegisterRequest struct {
	TenantID uint   `json:"tenantId"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
type RegisterResponse struct {
	UserID       uint   `json:"userId"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ProfileUpdateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PasswordResetRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type AuthHandler struct {
	DB     *gorm.DB
	Secret string
}

func NewAuthHandler(db *gorm.DB, secret string) *AuthHandler {
	return &AuthHandler{DB: db, Secret: secret}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, _ := auth.GenerateAccessToken(user.ID, user.TenantID, user.Role, h.Secret)
	refreshToken, _ := auth.GenerateRefreshToken(user.ID, user.TenantID, user.Role, h.Secret)

	// Only set Secure=true if this request is over HTTPS.
	isSecure := r.TLS != nil

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	resp := LoginResponse{
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		ForcePasswordChange: user.ForcePasswordChange,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	claims, err := auth.ParseRefreshToken(req.RefreshToken, h.Secret)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	accessToken, err := auth.GenerateAccessToken(claims.UserID, claims.TenantID, claims.Role, h.Secret)
	if err != nil {
		http.Error(w, "Token error", http.StatusInternalServerError)
		return
	}
	resp := RefreshResponse{AccessToken: accessToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" || req.TenantID == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	var existing models.User
	if err := h.DB.
		Where("email = ? AND tenant_id = ?", req.Email, req.TenantID).
		First(&existing).Error; err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Password error", http.StatusInternalServerError)
		return
	}
	user := models.User{
		TenantID:     req.TenantID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashed),
		Role:         req.Role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.DB.Create(&user).Error; err != nil {
		http.Error(w, "Create user failed", http.StatusInternalServerError)
		return
	}
	accessToken, _ := auth.GenerateAccessToken(user.ID, user.TenantID, user.Role, h.Secret)
	refreshToken, _ := auth.GenerateRefreshToken(user.ID, user.TenantID, user.Role, h.Secret)
	resp := RegisterResponse{
		UserID:       user.ID,
		Name:         user.Name,
		Email:        user.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Profile Endpoints ---

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	var payload ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	user.Name = payload.Name
	user.Email = payload.Email
	user.UpdatedAt = time.Now()
	if err := h.DB.Save(&user).Error; err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 8 {
		http.Error(w, "New password must be ≥8 characters", http.StatusBadRequest)
		return
	}
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "Current password incorrect", http.StatusBadRequest)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Password error", http.StatusInternalServerError)
		return
	}
	user.PasswordHash = string(hashed)
	user.ForcePasswordChange = false
	user.UpdatedAt = time.Now()
	if err := h.DB.Save(&user).Error; err != nil {
		http.Error(w, "Reset failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successful"})
}
