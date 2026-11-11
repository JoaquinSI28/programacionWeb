# TP5 — registro de implementación

Fecha: 11 de noviembre de 2026.
Rama: tp5, creada desde el TP4 terminado en 392b6bf. Las ramas tp3 y tp4 se conservan.

## Fuentes y alcance

- tp5.pdf, Trabajo de Cursada, página 3: componentes layout/lista/formulario en views, GET que consulta y renderiza HTML, POST que parsea/crea/redirige, generación y ejecución automatizadas.
- Consignas y reglas de entrega.md, TP5: rama tp5, README/documentación, sin archivos .js, preparación y limpieza de contenedores y volumen mediante comando documentado.
- 09-view-static.pdf, páginas físicas 3–5: SSR y recargas completas; 15–20: componentes y generación Go; 22: CSS estático/StripPrefix; 23–37: templ y sqlc; 38–39: dependencias.
- 07-logic.pdf, páginas físicas 13–15: responsabilidades y validaciones.
- Documentación oficial templ para composición y misma versión de generador/módulo: https://templ.guide/syntax-and-usage/template-composition/ y https://templ.guide/quick-start/installation/ .

Se usa el TP4 real terminado, no la instantánea tp3 de la vault. El ejemplo didáctico de productos/SQLite no cambia Bookfly/PostgreSQL.

## Transformación

- El HTML del TP4 se divide en views/layout.templ, entity_form.templ, entity_list.templ e index.templ.
- Cada lista recibe entidades tipadas de logic. PageData reúne listas, referencias y valores seguros del formulario. Las vistas no conocen sqlc.
- API/web.go consulta logic → repository → sqlc. Se mantienen las validaciones existentes, las claves foráneas y los contratos de repositorio.
- GET / renderiza la página completa. POST /libros, /autores, /editoriales y /usuarios recibe application/x-www-form-urlencoded, parsea con ParseForm, crea y redirige a / con 303 See Other.
- Los errores devuelven HTML con 400/404/409/500 según corresponda. Se conservan campos seguros; no se devuelve la contraseña.
- El borrado que ya existía en TP4 se conserva mediante POST /entidades/{id}/eliminar y PRG. Es una decisión de adaptación, no un requisito adicional atribuido al TP5. GET nunca elimina. TP6 será el paso a DELETE/HTMX sin recarga.
- Se elimina el cliente app.js, su HTML estático y los handlers JSON en esta rama. La prueba de navegador JavaScript del TP4 permanece en la rama tp4; la entrega tp5 no incluye .js/.cjs/.mjs ni etiquetas script o atributos de eventos JavaScript.
- Se conserva el CSS del TP4, adaptando los formularios de borrado.

## Herramientas y ejecución

Go 1.26.7 en el módulo y contenedor 1.26, PostgreSQL 15, lib/pq 1.12.3. Se agrega templ v0.3.1100, observado en el registro oficial de módulos; CLI e importación fijados a la misma versión. sqlc se fija a 1.31.1, versión instalada comprobada, evitando latest en ese generador.

make generate ejecuta sqlc y templ en contenedores. El Dockerfile también genera desde fuentes antes de compilar, y excluye los paquetes generados del contexto de entrada. make run-app prepara, arranca, espera salud y limpia al salir con Ctrl+C. make up y make stop permiten trabajar conservando datos. El nombre Compose predeterminado bookfly-tp5 separa el volumen del TP4; Docker asigna el puerto web y el script imprime la dirección efectiva. PostgreSQL usa la red interna de Compose.

## Decisiones conservadas

No se agrega autenticación, gestión de lecturas, edición de entidades ni reglas de unicidad nuevas. La unicidad de editoriales quedó explícitamente diferida por el usuario. Las altas y actualizaciones guardan PBKDF2-SHA256 con 600.000 iteraciones y sal aleatoria mediante la biblioteca estándar de Go. La aplicación migra la columna anterior al iniciar. La clase de autenticación fundamenta proteger las contraseñas con un hash lento; PBKDF2 es la elección del proyecto.

## Validación ejecutada

Se usaron proyectos Compose de prueba `bookfly-tp5-check` (puertos 18082/15434) y `bookfly-tp5-clean` (18083/15435), separados de la instancia y datos personales del TP4.

- Generación sqlc y templ: correcta, con las versiones fijadas.
- Compilación y arranque de aplicación/PostgreSQL saludables: correctos.
- `go test ./... -v` en contenedor contra PostgreSQL real: aprobado. Incluye TestHTMLFormsAndPRG y sus subcasos (validaciones del servidor, rutas y borrado), TestUsuario_CRUD y TestLibro_CRUD.
- `make test-hurl`: 18 peticiones aprobadas, 0 fallos; preparación y limpieza automáticas verificadas.
- Prueba de navegador Chrome con JavaScript desactivado: HTML inicial, cuatro altas, selectores de relaciones, POST 303 seguido de GET 200, recarga sin duplicados, conflictos en HTML, conservación de campos seguros, contraseña ausente y cuatro borrados: aprobada.
- Capturas de escritorio y celular revisadas. Ancho móvil de 390 px, sin desborde horizontal (scrollWidth=390). La captura móvil final se tomó desde una ventana creada directamente a ese ancho.
- Copia de fuentes sin db/sqlc ni *_templ.go: `make run-app` generó, compiló y sirvió HTTP 200. Se simuló Ctrl+C y se comprobó que quedaron 0 contenedores y 0 volúmenes de ese proyecto.
- No hay archivos propios .js, .cjs ni .mjs, ni etiquetas script/atributos JavaScript en las vistas.
- `git diff --check` y `bash -n scripts/run-app.sh`: correctos.

Las primeras verificaciones detectaron un import duplicado en una plantilla y una ruta comodín que respondía 405 para direcciones inexistentes; se corrigieron y las pruebas de la versión final pasaron. La primera ejecución combinada de make test-all aprobó Go pero detectó que Hurl enviaba comillas literales en el caso de campo vacío; se corrigió el caso de prueba y make test-hurl pasó completo. No se presenta aquella primera ejecución combinada como una ejecución exitosa.

La prueba de navegador es evidencia de desarrollo conservada fuera de la entrega, para que la rama TP5 no incluya JavaScript propio. Las pruebas reproducibles entregadas son las de Go y Hurl. Safari no se automatizó en esta etapa; queda la comprobación manual del usuario.

## Cierre

La revisión final incorporó el límite de 255 caracteres, la proyección pública de usuarios, la migración de contraseñas y el control de salud independiente. La rama tp5 conserva el flujo SSR de la cursada.

La ejecución final de `make test-all` pasó las pruebas de Go y las 18 peticiones Hurl. Las pruebas de regresión cubren contraseñas de 255 y 256 caracteres, respuesta vacía 204 de salud y migración repetible de dos usuarios con la misma clave y sales distintas. También pasó `go vet ./...`.
