package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"bookfly.com/logic"
)

func mensajeValidacion(err error) string {
	var validacion *logic.ErrorValidacion
	if errors.As(err, &validacion) {
		return validacion.Mensaje
	}
	return "Datos inválidos"
}

func responderConflicto(w http.ResponseWriter, err error) bool {
	if errors.Is(err, logic.ErrDuplicado) || errors.Is(err, logic.ErrReferencia) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
		return true
	}
	return false
}

func metodoNoPermitido(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") == "" {
		w.Header().Set("Allow", "GET, POST")
	} else {
		w.Header().Set("Allow", "GET, PUT, DELETE")
	}
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "Método no permitido"})
}

func decodeJSON(r *http.Request, request any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(request); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("Cuerpo JSON inválido")
	}
	return nil
}
