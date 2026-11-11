# Bookfly — TP5: renderizado del servidor con templ

Registro de libros desarrollado para Programación Web, UNICEN. La rama `tp5` continúa el TP4 terminado en `392b6bf`, conservando PostgreSQL, sqlc, las entidades y las validaciones existentes.

En este TP, Go obtiene los datos y templ produce el HTML completo. Los formularios funcionan con JavaScript desactivado: envían POST, el servidor guarda y responde `303 See Other` hacia `/`; el navegador hace un GET y muestra la lista actualizada. Es el patrón **Post/Redirect/Get (PRG)**.

## Requisitos

Docker Desktop iniciado, Docker Compose y make. Go, sqlc, templ y Hurl se ejecutan dentro de contenedores; no hace falta instalarlos en el equipo.

Versiones: Go 1.26.7 en el módulo (imagen Go 1.26), PostgreSQL 15, sqlc 1.31.1 y templ 0.3.1100. El módulo templ y su generador usan la misma versión. Se conserva lib/pq 1.12.3.

## Comando de entrega

```bash
make run-app
```

Este comando:

1. Limpia contenedores y volumen anteriores **del proyecto `bookfly-tp5`**.
2. Genera Go desde SQL (`sqlc`) y desde componentes (`templ`).
3. Compila la aplicación y levanta PostgreSQL y el servidor, esperando sus comprobaciones de salud.
4. Muestra los registros del servidor. Abrir la dirección que aparece junto a **Bookfly disponible en**.
5. Al salir con **Ctrl+C**, detiene los servicios y elimina el volumen de ese proyecto, incluidos sus datos, como requiere el arranque de entrega.

Docker asigna un puerto web local disponible. PostgreSQL se comunica con la aplicación por la red interna de Compose. El arranque de entrega funciona aunque otro proyecto use 8080 y 5432. Los datos del volumen del TP5 se eliminan al cerrar este modo de ejecución.

## Desarrollo conservando los datos

```bash
make up
```

Genera, compila e inicia los servicios en segundo plano. Para consultar la dirección web, ejecutá `docker compose -p bookfly-tp5 port api 8080` y abrí esa dirección con HTTP. Para detenerlos conservando el volumen:

```bash
make stop
```

Para eliminarlos junto con los datos:

```bash
make down
```

Usar estos targets mantiene el nombre de proyecto del TP5. Si se personaliza `COMPOSE_PROJECT_NAME`, usar el mismo valor al arrancar y detener. Se admiten `.env` local o variables de entorno `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` y `APP_PORT` (opcional para fijar el puerto web). `.env` no se versiona. Los valores de desarrollo predeterminados son base `prueba`, usuario `postgres`, contraseña `postgres`.

## Usar la aplicación

1. Agregar un autor y una editorial.
2. Crear un libro eligiendo esas referencias, fecha de lanzamiento y páginas.
3. Crear usuarios desde su sección.
Las contraseñas admiten entre 3 y 255 caracteres. El repositorio guarda un hash PBKDF2-SHA256 con sal aleatoria. Al arrancar sobre un volumen anterior, el servidor migra los usuarios existentes en una transacción. El listado consulta únicamente ID, nombre y correo.

Para eliminar registros, usá los botones de cada sección. El borrado también usa un formulario POST y PRG, para conservar la operación del TP4 sin JavaScript. Un autor/editorial con libros relacionados no se puede eliminar hasta borrar esos libros.

Cada envío correcto produce una navegación completa. Recargar después de la redirección repite el GET, no el POST. Si los datos son inválidos, el servidor devuelve la página con el error y conserva los campos seguros del formulario; la contraseña debe volver a ingresarse. La base sigue permitiendo editoriales con igual nombre/sede, decisión pendiente del dominio.

## Relación con la consigna y la teoría

Fuente: `tp5.pdf`, sección «Trabajo de Cursada: Renderizado del Lado del Servidor con Templ», página 3. Productos y SQLite son el ejemplo didáctico, no una migración del dominio Bookfly.

| Requisito | Implementación |
| --- | --- |
| Componentes en views | `layout.templ`, `entity_list.templ`, `entity_form.templ`, `index.templ`. |
| GET / consulta y renderiza HTML | `API/web.go`: usa logic → repository → consultas sqlc y renderiza `views.IndexPage`. |
| POST parsea, crea y redirige | `r.ParseForm`, validaciones existentes, Create de sqlc a través del repositorio y `http.Redirect` con 303. |
| Generación y ejecución automatizadas | Makefile, `scripts/run-app.sh`, Compose y Dockerfile. |
| Entrega sin .js | No hay archivos .js, .cjs ni .mjs propios; tampoco scripts ni eventos JavaScript en las vistas. |

Teoría: `09-view-static.pdf`, páginas físicas 3–5 (SSR y recargas completas), 15–20 (templ y flujo), 22 (CSS estático con StripPrefix), 23–37 (componentes y sqlc), 38–39 (dependencias). `07-logic.pdf`, páginas 13–15, fundamenta conservar las capas y validaciones. La composición se contrastó con https://templ.guide/syntax-and-usage/template-composition/ .

Los formularios de borrado, los mapas de nombres de autores/editoriales y la conservación de campos ante errores son decisiones para adaptar lo que ya tenía el TP4; la sección de cursada exige principalmente listado y alta. HTMX queda para TP6.

## Estructura

- `API/main.go`: conexión y servidor.
- `API/web.go`: rutas HTML, formularios, errores y PRG.
- `API/static/styles.css`: estilos del TP4, adaptados a formularios de borrado.
- `views/`: componentes templ y datos de presentación.
- `logic/`: entidades, contratos y validaciones existentes.
- `repository/`: implementaciones con sqlc existentes.
- `db/schema/` y `db/queries/`: PostgreSQL y SQL existentes.
- `db/sqlc/` y `views/*_templ.go`: Go generado; no se versiona.
- `API/web_test.go`: pruebas HTTP con base real.
- `requests.hurl`: recorrido HTML reproducible.

Los handlers JSON, HTML estático y app.js del TP4 quedan en su rama histórica `tp4`. El Dockerfile genera sqlc y templ desde las fuentes también durante la compilación, sin depender de archivos generados previamente en la máquina.

## Pruebas automatizadas

```bash
make test
make test-hurl
# O ejecutar ambas:
make test-all
```

Las pruebas limpian contenedores y volumen de `bookfly-tp5` antes y después, incluso si fallan. Se pueden aislar con otro `COMPOSE_PROJECT_NAME` y otros puertos igual que el arranque. `make test` genera los paquetes y ejecuta `go test ./...` dentro de Docker contra PostgreSQL real. `make test-hurl` comprueba HTML, formularios, 303/Location, GET, errores, borrado y ausencia de JavaScript con Hurl.

La prueba HTTP verifica además que recargar no duplique altas, que se escape el contenido HTML del usuario, que no se devuelvan contraseñas, que fallen números/fechas/referencias inválidas y que GET no permita borrar.

## Comprobación manual

- Ver el código fuente de la página: las listas y sus registros deben estar en el HTML recibido.
- Desactivar JavaScript y crear autor, editorial, usuario y libro.
- En Red/Network, observar POST → 303 → GET / → 200 al guardar.
- Recargar y confirmar persistencia sin duplicados.
- Probar usuario repetido y borrado de autor con libros asociados: se debe mostrar el error en HTML.
- Borrar el libro y después sus referencias.
- Revisar escritorio y ancho móvil de 390 px.

Ver `docs/TP5.md` para el registro de implementación, pruebas ejecutadas y pendientes.

El control de salud consulta una constante en PostgreSQL mediante sqlc y devuelve HTTP 204 en `GET /salud`.
