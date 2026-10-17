package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"bookfly.com/logic"
)

type editorialResponse struct {
	ID     int32  `json:"id"`
	Nombre string `json:"nombre"`
	Sede   string `json:"sede"`
}

type editorialRequest struct {
	Nombre string `json:"nombre"`
	Sede   string `json:"sede"`
}

type editorialesHandler struct {
	repo logic.EditorialRepository
}

func editorialHandler(repo logic.EditorialRepository) http.Handler {
	return &editorialesHandler{repo: repo}
}

func (h *editorialesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") == "" && (r.Method == http.MethodPut || r.Method == http.MethodDelete) ||
		r.PathValue("id") != "" && r.Method == http.MethodPost {
		metodoNoPermitido(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if r.PathValue("id") == "" {
			h.getEditoriales(w, r)
			return
		}
		h.getEditorial(w, r)
	case http.MethodPost:
		h.createEditorial(w, r)
	case http.MethodPut:
		h.updateEditorial(w, r)
	case http.MethodDelete:
		h.deleteEditorial(w, r)
	default:
		metodoNoPermitido(w, r)
	}
}

func (h *editorialesHandler) createEditorial(w http.ResponseWriter, r *http.Request) {
	request, err := decodeEditorialRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	editorial, err := logic.CreateEditorial(r.Context(), h.repo, request.Nombre, request.Sede)
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		if responderConflicto(w, err) {
			return
		}
		log.Printf("Error al crear editorial: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	writeJSON(w, http.StatusCreated, editorialToResponse(editorial))
}

func (h *editorialesHandler) getEditorial(w http.ResponseWriter, r *http.Request) {
	id, err := parseEditorialID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de editorial inválido"})
		return
	}

	editorial, err := logic.GetEditorial(r.Context(), h.repo, id)
	if errors.Is(err, logic.ErrEditorialNoEncontrada) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Editorial no encontrada"})
		return
	}
	if err != nil {
		log.Printf("Error al buscar editorial %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	writeJSON(w, http.StatusOK, editorialToResponse(editorial))
}

func (h *editorialesHandler) getEditoriales(w http.ResponseWriter, r *http.Request) {
	editoriales, err := logic.GetEditoriales(r.Context(), h.repo)
	if err != nil {
		log.Printf("Error al obtener editoriales: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	responses := make([]editorialResponse, 0, len(editoriales))
	for _, editorial := range editoriales {
		responses = append(responses, editorialToResponse(editorial))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (h *editorialesHandler) updateEditorial(w http.ResponseWriter, r *http.Request) {
	id, err := parseEditorialID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de editorial inválido"})
		return
	}

	request, err := decodeEditorialRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	editorial, err := logic.UpdateEditorial(r.Context(), h.repo, id, request.Nombre, request.Sede)
	if errors.Is(err, logic.ErrEditorialNoEncontrada) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Editorial no encontrada"})
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
		log.Printf("Error al actualizar editorial %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	writeJSON(w, http.StatusOK, editorialToResponse(editorial))
}

func (h *editorialesHandler) deleteEditorial(w http.ResponseWriter, r *http.Request) {
	id, err := parseEditorialID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de editorial inválido"})
		return
	}

	if err := logic.DeleteEditorial(r.Context(), h.repo, id); errors.Is(err, logic.ErrEditorialNoEncontrada) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Editorial no encontrada"})
		return
	} else if err != nil {
		if responderConflicto(w, err) {
			return
		}
		log.Printf("Error al eliminar editorial %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseEditorialID(r *http.Request) (int32, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("ID de editorial inválido")
	}
	return int32(id), nil
}

func decodeEditorialRequest(r *http.Request) (editorialRequest, error) {
	var request editorialRequest
	if err := decodeJSON(r, &request); err != nil {
		return editorialRequest{}, errors.New("Cuerpo JSON inválido")
	}
	return request, nil
}

func editorialToResponse(editorial *logic.Editorial) editorialResponse {
	return editorialResponse{
		ID:     editorial.IDEditorial,
		Nombre: editorial.Nombre,
		Sede:   editorial.Sede,
	}
}
