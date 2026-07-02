package handlers

import (
	"net/http"

	"nutricionist/auth"
)

func (h *Handler) UpdatePerfil(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body struct {
		Nombre    *string `json:"nombre"`
		Apellidos *string `json:"apellidos"`
		Cedula    *string `json:"cedula"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	_, err := h.db.ExecContext(r.Context(), `
		UPDATE nutricionistas SET
			nombre    = COALESCE(?, nombre),
			apellidos = COALESCE(?, apellidos),
			cedula    = COALESCE(?, cedula)
		WHERE id = ?`,
		body.Nombre, body.Apellidos, body.Cedula, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error actualizando perfil")
		return
	}

	var result struct {
		ID        string  `json:"id"`
		Email     string  `json:"email"`
		Nombre    string  `json:"nombre"`
		Apellidos string  `json:"apellidos"`
		Cedula    *string `json:"cedula"`
	}
	h.db.QueryRowContext(r.Context(),
		"SELECT id,email,nombre,apellidos,cedula FROM nutricionistas WHERE id=?", userID).
		Scan(&result.ID, &result.Email, &result.Nombre, &result.Apellidos, &result.Cedula)

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body struct {
		PasswordActual string `json:"passwordActual"`
		PasswordNueva  string `json:"passwordNuevo"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if len(body.PasswordNueva) < 8 {
		writeError(w, http.StatusBadRequest, "la contrasena debe tener al menos 8 caracteres")
		return
	}

	var hash string
	if err := h.db.QueryRowContext(r.Context(),
		"SELECT password FROM nutricionistas WHERE id=?", userID).Scan(&hash); err != nil {
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if !auth.CheckPassword(hash, body.PasswordActual) {
		writeError(w, http.StatusUnauthorized, "contrasena actual incorrecta")
		return
	}

	newHash, err := auth.HashPassword(body.PasswordNueva)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error procesando contrasena")
		return
	}
	h.db.ExecContext(r.Context(),
		"UPDATE nutricionistas SET password=? WHERE id=?", newHash, userID)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
