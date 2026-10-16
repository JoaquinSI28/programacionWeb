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

type libroResponse struct {
	IDLibro          int32  `json:"id"`
	Titulo           string `json:"titulo"`
	IDAutor          int32  `json:"id_autor"`
	IDEditorial      int32  `json:"id_editorial"`
	FechaLanzamiento string `json:"fecha_lanzamiento"`
	CantidadPaginas  int32  `json:"cantidad_paginas"`
}

type libroRequest struct {
	Titulo           string `json:"titulo"`
	IDAutor          int32  `json:"id_autor"`
	IDEditorial      int32  `json:"id_editorial"`
	FechaLanzamiento string `json:"fecha_lanzamiento"`
	CantidadPaginas  int32  `json:"cantidad_paginas"`
}

type librosHandler struct {
	repo logic.LibroRepository
}

func libroHandler(repo logic.LibroRepository) http.Handler {
	return &librosHandler{repo: repo}
}

func (h *librosHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.PathValue("id") == "" {
			h.getLibros(w, r)
			return
		}
		h.getLibro(w, r)
	case http.MethodPost:
		h.createLibro(w, r)
	case http.MethodPut:
		h.updateLibro(w, r)
	case http.MethodDelete:
		h.deleteLibro(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "Método no permitido"})
	}
}

func (h *librosHandler) createLibro(w http.ResponseWriter, r *http.Request) {
	request, fecha, err := decodeLibroRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	libro, err := logic.CreateLibro(r.Context(), h.repo, request.IDAutor, request.Titulo, request.IDEditorial, fecha, request.CantidadPaginas)
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		log.Printf("Error al crear libro: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	writeJSON(w, http.StatusCreated, libroToResponse(libro))
}

func (h *librosHandler) getLibro(w http.ResponseWriter, r *http.Request) {
	id, err := parseLibroID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de libro inválido"})
		return
	}

	libro, err := logic.GetLibro(r.Context(), h.repo, id)
	if errors.Is(err, logic.ErrLibroNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Libro no encontrado"})
		return
	}
	if err != nil {
		log.Printf("Error al buscar libro %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	writeJSON(w, http.StatusOK, libroToResponse(libro))
}

func (h *librosHandler) getLibros(w http.ResponseWriter, r *http.Request) {
	libros, err := logic.GetLibros(r.Context(), h.repo)
	if err != nil {
		log.Printf("Error al obtener libros: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	responses := make([]libroResponse, 0, len(libros))
	for _, libro := range libros {
		responses = append(responses, libroToResponse(libro))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (h *librosHandler) updateLibro(w http.ResponseWriter, r *http.Request) {
	id, err := parseLibroID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de libro inválido"})
		return
	}

	request, fecha, err := decodeLibroRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	libro, err := logic.UpdateLibro(r.Context(), h.repo, id, request.IDAutor, request.Titulo, request.IDEditorial, fecha, request.CantidadPaginas)
	if errors.Is(err, logic.ErrLibroNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Libro no encontrado"})
		return
	}
	if err != nil {
		if errors.Is(err, logic.ErrValidacion) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: mensajeValidacion(err)})
			return
		}
		log.Printf("Error al actualizar libro %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	writeJSON(w, http.StatusOK, libroToResponse(libro))
}

func (h *librosHandler) deleteLibro(w http.ResponseWriter, r *http.Request) {
	id, err := parseLibroID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "ID de libro inválido"})
		return
	}

	if err := logic.DeleteLibro(r.Context(), h.repo, id); errors.Is(err, logic.ErrLibroNoEncontrado) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "Libro no encontrado"})
		return
	} else if err != nil {
		log.Printf("Error al eliminar libro %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Error interno del servidor"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeLibroRequest(r *http.Request) (libroRequest, logic.Date, error) {
	var request libroRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return libroRequest{}, logic.Date{}, errors.New("Cuerpo JSON inválido")
	}

	fecha, err := time.Parse("2006-01-02", request.FechaLanzamiento)
	if err != nil {
		return libroRequest{}, logic.Date{}, errors.New("La fecha de lanzamiento debe tener formato YYYY-MM-DD")
	}
	return request, fecha, nil
}

func parseLibroID(r *http.Request) (int32, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("ID de libro inválido")
	}
	return int32(id), nil
}

func libroToResponse(libro *logic.Libro) libroResponse {
	return libroResponse{
		IDLibro:          libro.IDLibro,
		Titulo:           libro.Titulo,
		IDAutor:          libro.IDAutor,
		IDEditorial:      libro.IDEditorial,
		FechaLanzamiento: libro.FechaLanzamiento.Format("2006-01-02"),
		CantidadPaginas:  libro.CantidadPaginas,
	}
}
