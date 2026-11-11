# TP4 — registro de implementación y verificación

Fecha: 11 de noviembre de 2026.

## Base y alcance

Rama `tp4`, creada desde `tp3` en `4e70149de054ec058f2c6d617e9511cf06da3826`. El siguiente TP deberá crear `tp5` desde el commit final de `tp4`, sin modificar la rama entregada. No se publicó ninguna rama durante esta implementación.

Fuente funcional: guía oficial `tp4.pdf`, Trabajo de Cursada, páginas 2–3. Material de apoyo: `08-view.pdf`, páginas físicas 15 (origen/CORS), 19–27 (DOM y eventos), 59 (formularios), 71–78 (Flexbox y Grid).

Se implementaron los cuatro puntos de cursada: HTML con formularios y listas identificables, fetch de listas y creación JSON, borrado con delegación de eventos y estilos con Grid/Flexbox. Se cubren libros, autores, editoriales y usuarios, que ya tenían endpoints en el TP3. La relación Lee sigue en datos, como en la entrega anterior.

## Archivos

- `API/static/index.html`: estructura semántica, cuatro formularios y listas, navegación por secciones, etiquetas y mensajes accesibles.
- `API/static/app.js`: GET inicial, POST/DELETE, renderizado con nodos de texto, actualización de listas y selectores de referencias, errores y estados de carga.
- `API/static/styles.css`: diseño responsivo sin dependencias externas.
- `tests/tp4-browser.cjs`: prueba reproducible de navegador contra API y base reales.
- `README.md`: arranque, uso, correspondencia con consigna y pruebas.

No se modificaron el backend Go, sus consultas SQL, dependencias ni contenedores de la entrega. Los archivos estáticos nuevos se incluyen mediante el COPY ya existente en Dockerfile.

## Decisiones propias

- Servir frontend y API desde el mismo servidor para usar rutas relativas sin agregar CORS.
- Seleccionar autores/editoriales existentes para los libros, en lugar de pedir IDs manualmente.
- Conservar validaciones de negocio en Go; las restricciones HTML y el chequeo de espacios dan respuesta inmediata en el cliente.
- Usar textContent para mostrar entradas sin interpretar HTML.
- Deshabilitar botones durante las operaciones para evitar dobles envíos.
- Diferenciar un fallo del POST de un fallo de lectura posterior: si la creación se confirmó, limpiar el formulario e informar que el cambio ya fue guardado.

## Evidencia ejecutada

Entorno aislado Docker Compose `bookfly-tp4-check`, con contenedores y volumen propios. Puerto web local 18080 y PostgreSQL local 15432 para evitar interferencia con instancias personales.

1. Generación sqlc: correcta.
2. Compilación de imagen Go y arranque con API y PostgreSQL saludables: correctos. Un primer intento de compilación previo a generar sqlc falló por falta de `db/sqlc`; se generó y se repitió exitosamente, siguiendo el orden que ya usa `make up`.
3. Sintaxis de app.js y prueba de navegador: correcta con `node --check`.
4. Prueba de navegador ejecutada con Google Chrome, Playwright 1.62.1 y API/PostgreSQL reales: aprobada.
   - Cuatro listas inicialmente vacías.
   - Alta de las cuatro entidades y relaciones de libros.
   - IDs y páginas enviados como números JSON.
   - Creación sin navegación del documento y persistencia tras recarga.
   - Texto con apariencia HTML mostrado como texto sin ejecutar código.
   - Contraseña ausente de la lista y campo limpiado al crear.
   - Rechazo visible de usuario duplicado, conservando formulario.
   - Rechazo de campos con solo espacios antes del POST.
   - Rechazo de borrado de autor con libro asociado y conservación del registro.
   - Borrado de las cuatro entidades y botones funcionales tras regenerar listas.
   - Fallo de red distinguido de lista vacía.
   - Creación confirmada seguida de fallo de consulta: mensaje de cambio guardado; recarga confirma persistencia.
   - Vista móvil de 390 px sin desborde horizontal; revisión visual de capturas de escritorio y celular.
   - Sin errores JavaScript no capturados.
5. `requests.hurl` original: 40 peticiones aprobadas, 0 fallos.
6. `git diff --check`: correcto.

La prueba de navegador y la instancia aislada son de desarrollo; la aplicación sigue arrancando con Docker/Compose/make, sin Node en ejecución. No se ejecutó la suite de persistencia Go en esta etapa: no se modificaron esas capas y se validó la persistencia a través del navegador y Hurl.

## Continuidad

TP4 implementado y validado. TP5 pendiente: deberá transformar esta interfaz en componentes templ y formularios clásicos, conservando el dominio y PostgreSQL. Publicación en GitHub pendiente de indicación del usuario.
