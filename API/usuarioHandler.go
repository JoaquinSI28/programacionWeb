package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"bookfly.com/logic"
)

type usuarioResponse struct {
	ID     int32  `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func usuarioHandler(repo logic.UsuarioRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de usuario inválido"})
			return
		}

		usuario, err := logic.GetUsuario(r.Context(), repo, int32(id))
		if errors.Is(err, logic.ErrUsuarioNoEncontrado) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "Usuario no encontrado"})
			return
		}
		if err != nil {
			log.Printf("Error al buscar usuario %d: %v", id, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
			return
		}

		writeJSON(w, http.StatusOK, usuarioResponse{
			ID:     int32(id),
			Nombre: usuario.Nombre,
			Email:  usuario.Email,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
