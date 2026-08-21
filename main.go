package main

import (
  "fmt"
  "net/http"
)

func main() {
  // 1. Define el directorio que contiene los archivos estáticos.
  staticDir := "./static"

  // 2. Crea un manejador (handler) de servidor de archivos.
  // http.Dir convierte la ruta del directorio en un sistema de archivos HTTP.
  // http.FileServer crea un manejador que sirve archivos desde ese sistema.
  // ¡Automáticamente sirve index.html para directorios!
  fileServer := http.FileServer(http.Dir(staticDir))

	// 2. Registra un manejador (handler) para la ruta raíz "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 3. Establece la cabecera Content-Type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 4. Escribe el HTML en la respuesta
		fmt.Fprint(w, htmlContent)
	})
  // 5. Define el puerto y muestra un mensaje
	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)

	// 6. Inicia el servidor HTTP
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}
