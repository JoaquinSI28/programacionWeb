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

type createUsuarioRequest struct {
	Nombre     string `json:"nombre"`
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

type usuariosHandler struct {
	repo logic.UsuarioRepository
}

func usuarioHandler(repo logic.UsuarioRepository) http.Handler {
	return &usuariosHandler{repo: repo}
}

func (h *usuariosHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") == "" && (r.Method == http.MethodPut || r.Method == http.MethodDelete) ||
		r.PathValue("id") != "" && r.Method == http.MethodPost {
		metodoNoPermitido(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if r.PathValue("id") == "" {
			h.getUsuarios(w, r)
			return
		}
		h.getUsuario(w, r)
	case http.MethodPost:
		h.createUsuario(w, r)
	case http.MethodPut:
		h.updateUsuario(w, r)
	case http.MethodDelete:
		h.deleteUsuario(w, r)
	default:
		metodoNoPermitido(w, r)
	}
}

func (h *usuariosHandler) createUsuario(w http.ResponseWriter, r *http.Request) {
	var request createUsuarioRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Cuerpo JSON inválido"})
		return
	}

	usuario, err := logic.CreateUsuario(r.Context(), h.repo, request.Nombre, request.Email, request.Contrasena)
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		if responderConflicto(w, err) {
			return
		}
		log.Printf("Error al crear usuario: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	writeJSON(w, http.StatusCreated, usuarioResponse{
		ID:     usuario.ID,
		Nombre: usuario.Nombre,
		Email:  usuario.Email,
	})
}

func (h *usuariosHandler) getUsuario(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de usuario inválido"})
		return
	}

	usuario, err := logic.GetUsuario(r.Context(), h.repo, int32(id))
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

func (h *usuariosHandler) updateUsuario(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de usuario inválido"})
		return
	}

	var request createUsuarioRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Cuerpo JSON inválido"})
		return
	}

	usuario, err := logic.UpdateUsuario(r.Context(), h.repo, int32(id), request.Nombre, request.Email, request.Contrasena)
	if errors.Is(err, logic.ErrUsuarioNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Usuario no encontrado"})
		return
	}
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		if responderConflicto(w, err) {
			return
		}
		log.Printf("Error al actualizar usuario %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	writeJSON(w, http.StatusOK, usuarioResponse{
		ID:     usuario.ID,
		Nombre: usuario.Nombre,
		Email:  usuario.Email,
	})
}

func (h *usuariosHandler) getUsuarios(w http.ResponseWriter, r *http.Request) {
	usuarios, err := logic.GetUsuarios(r.Context(), h.repo)
	if err != nil {
		log.Printf("Error al obtener usuarios: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	responses := make([]usuarioResponse, 0, len(usuarios))

	for _, usuario := range usuarios {
		responses = append(responses, usuarioResponse{
			ID:     usuario.ID,
			Nombre: usuario.Nombre,
			Email:  usuario.Email,
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

func (h *usuariosHandler) deleteUsuario(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de usuario inválido"})
		return
	}

	err = logic.DeleteUsuario(r.Context(), h.repo, int32(id))
	if errors.Is(err, logic.ErrUsuarioNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Usuario no encontrado"})
		return
	}
	if err != nil {
		if responderConflicto(w, err) {
			return
		}
		log.Printf("Error al eliminar usuario %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
