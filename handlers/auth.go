package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"nutricionist/auth"
	"nutricionist/db"
)

// AuthInput represents login request body
type AuthInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterInput represents registration request body
type RegisterInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Nombre    string `json:"nombre"`
	Apellidos string `json:"apellidos"`
	Cedula    string `json:"cedula"`
}

// RefreshInput represents token refresh request body
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse represents authentication response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// UserProfile represents nutricionist profile response
type UserProfile struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Nombre    string    `json:"nombre"`
	Apellidos string    `json:"apellidos"`
	Cedula    string    `json:"cedula"`
	CreatedAt time.Time `json:"created_at"`
}

// Login handles POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	
	if !auth.CheckRateLimit(ip) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "demasiados intentos"})
		return
	}

	var input AuthInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "json inválido")
		return
	}

	var userID, passwordHash string
	err := h.db.QueryRow("SELECT id, password FROM nutricionistas WHERE email = ?", strings.ToLower(strings.TrimSpace(input.Email))).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		auth.RecordFailedAttempt(ip)
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}

	if !auth.CheckPassword(passwordHash, input.Password) {
		auth.RecordFailedAttempt(ip)
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error generando tokens")
		return
	}

	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	})
}

// Registro handles POST /api/auth/registro
func (h *Handler) Registro(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "json inválido")
		return
	}

	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Password = strings.TrimSpace(input.Password)
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.Apellidos = strings.TrimSpace(input.Apellidos)
	input.Cedula = strings.TrimSpace(input.Cedula)

	if input.Email == "" || input.Password == "" {
		writeError(w, http.StatusBadRequest, "email y password son requeridos")
		return
	}
	if len(input.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password debe tener al menos 8 caracteres")
		return
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error procesando password")
		return
	}

	id := db.GenerateID()
	_, err = h.db.Exec(
		"INSERT INTO nutricionistas (id, email, password, nombre, apellidos, cedula, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, input.Email, passwordHash, input.Nombre, input.Apellidos, input.Cedula, time.Now(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			writeError(w, http.StatusConflict, "email ya registrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "error creando usuario")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":    id,
		"email": input.Email,
		"nombre": input.Nombre,
	})
}

// RefreshToken handles POST /api/auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "json inválido")
		return
	}

	userID, err := auth.ValidateToken(input.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "token inválido o expirado")
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error generando tokens")
		return
	}

	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	})
}

// GetMe handles GET /api/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	var profile UserProfile
	err := h.db.QueryRow(
		"SELECT id, email, nombre, apellidos, cedula, created_at FROM nutricionistas WHERE id = ?",
		userID,
	).Scan(&profile.ID, &profile.Email, &profile.Nombre, &profile.Apellidos, &profile.Cedula, &profile.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}
