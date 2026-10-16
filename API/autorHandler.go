package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"bookfly.com/logic"
)

type autorResponse struct {
	ID              int32  `json:"id"`
	Nombre          string `json:"nombre"`
	Pais            string `json:"pais"`
	FechaNacimiento string `json:"fecha_nacimiento"`
	Biografia       string `json:"biografia"`
}

type autorRequest struct {
	Nombre          string `json:"nombre"`
	Pais            string `json:"pais"`
	FechaNacimiento string `json:"fecha_nacimiento"`
	Biografia       string `json:"biografia"`
}

type autoresHandler struct {
	repo logic.AutorRepository
}

func autorHandler(repo logic.AutorRepository) http.Handler {
	return &autoresHandler{repo: repo}
}

func (h *autoresHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.PathValue("id") == "" {
			h.getAutores(w, r)
			return
		}
		h.getAutor(w, r)
	case http.MethodPost:
		h.createAutor(w, r)
	case http.MethodPut:
		h.updateAutor(w, r)
	case http.MethodDelete:
		h.deleteAutor(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "Método no permitido"})
	}
}

func (h *autoresHandler) createAutor(w http.ResponseWriter, r *http.Request) {
	request, fecha, err := decodeAutorRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	autor, err := logic.CreateAutor(r.Context(), h.repo, request.Nombre, request.Pais, fecha, request.Biografia)
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		log.Printf("Error al crear autor: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	writeJSON(w, http.StatusCreated, autorToResponse(autor))
}

func (h *autoresHandler) getAutor(w http.ResponseWriter, r *http.Request) {
	id, err := parseAutorID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de autor inválido"})
		return
	}

	autor, err := logic.GetAutor(r.Context(), h.repo, id)
	if errors.Is(err, logic.ErrAutorNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Autor no encontrado"})
		return
	}
	if err != nil {
		log.Printf("Error al buscar autor %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	writeJSON(w, http.StatusOK, autorToResponse(autor))
}

func (h *autoresHandler) getAutores(w http.ResponseWriter, r *http.Request) {
	autores, err := logic.GetAutores(r.Context(), h.repo)
	if err != nil {
		log.Printf("Error al obtener autores: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	responses := make([]autorResponse, 0, len(autores))
	for _, autor := range autores {
		responses = append(responses, autorToResponse(autor))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (h *autoresHandler) updateAutor(w http.ResponseWriter, r *http.Request) {
	id, err := parseAutorID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de autor inválido"})
		return
	}

	request, fecha, err := decodeAutorRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	autor, err := logic.UpdateAutor(r.Context(), h.repo, id, request.Nombre, request.Pais, fecha, request.Biografia)
	if errors.Is(err, logic.ErrAutorNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Autor no encontrado"})
		return
	}
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		log.Printf("Error al actualizar autor %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	writeJSON(w, http.StatusOK, autorToResponse(autor))
}

func (h *autoresHandler) deleteAutor(w http.ResponseWriter, r *http.Request) {
	id, err := parseAutorID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de autor inválido"})
		return
	}

	if err := logic.DeleteAutor(r.Context(), h.repo, id); errors.Is(err, logic.ErrAutorNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Autor no encontrado"})
		return
	} else if err != nil {
		log.Printf("Error al eliminar autor %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeAutorRequest(r *http.Request) (autorRequest, logic.Date, error) {
	var request autorRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return autorRequest{}, logic.Date{}, errors.New("Cuerpo JSON inválido")
	}

	fecha, err := time.Parse("2006-01-02", request.FechaNacimiento)
	if err != nil {
		return autorRequest{}, logic.Date{}, errors.New("La fecha de nacimiento debe tener formato YYYY-MM-DD")
	}
	return request, fecha, nil
}

func parseAutorID(r *http.Request) (int32, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("ID de autor inválido")
	}
	return int32(id), nil
}

func autorToResponse(autor *logic.Autor) autorResponse {
	return autorResponse{
		ID:              autor.IDAutor,
		Nombre:          autor.Nombre,
		Pais:            autor.Pais,
		FechaNacimiento: autor.FechaNacimiento.Format("2006-01-02"),
		Biografia:       autor.Biografia,
	}
}
