# TP6 — Interfaces reactivas con HTMX

La rama tp6 parte del commit f012f4f del TP5. El dominio, PostgreSQL, sqlc, las capas de lógica y repositorio y sus validaciones continúan desde esa entrega.

## Correspondencia con la materia

La fuente funcional es tp6.pdf, Trabajo de Cursada, páginas 2–3. La clase 10-view-dynamic.pdf desarrolla las respuestas HTML parciales, atributos HTMX y actualizaciones OOB. La composición de vistas continúa lo explicado en 09-view-static.pdf.

| Punto de cursada | Código |
| --- | --- |
| Cargar HTMX en el layout | views/layout.templ carga HTMX 2.0.11 desde CDN. |
| Formulario con hx-post, hx-target y hx-swap | views/entity_form.templ, en los cuatro formularios. |
| POST consulta y devuelve la lista actualizada | API/web.go y views/fragments.templ. |
| Botón con hx-delete y target del registro | views/entity_list.templ, target closest li y swap outerHTML. |
| DELETE devuelve 200 vacío | API/web.go. |
| Creación y borrado sin recarga | Verificación real en Chrome con una sola navegación completa. |

La guía muestra 1.9.10 como ejemplo. Se fijó 2.0.11 según la documentación oficial actual para configurar respuestas de error mediante HTML. Los estados 400, 404, 409 y 500 conservan su significado HTTP. Las cabeceras HX-Retarget y HX-Reswap dirigen el mensaje a una región accesible del formulario o del registro cuyo borrado falló.

## Decisiones de integración

La respuesta de creación contiene la lista correspondiente. Tras una operación correcta, HX-Trigger anuncia catalogoActualizado. Un GET a /estado devuelve contadores y selectores con hx-swap-oob. Esto mantiene disponibles los autores y editoriales nuevos sin reemplazar los demás campos que se están completando. Los selectores envían su selección actual para conservarla si la referencia sigue existiendo.

El borrado devuelve un cuerpo vacío, como pide la consigna. El mismo evento actualiza el estado auxiliar. Cada lista conserva un mensaje vacío que CSS muestra al eliminar su última entidad. Las peticiones se sincronizan en una cola mediante hx-sync para mantener coherencia entre operaciones y respuestas.

Los swaps OOB aparecen en los ejercicios didácticos y la clase. Se eligieron para los contadores y las referencias del dominio. La sección Trabajo de Cursada exige creación y borrado parciales.

## Rutas

| Método y ruta | Respuesta |
| --- | --- |
| GET / | Documento HTML completo. |
| POST /libros, /autores, /editoriales, /usuarios | Lista de la entidad en HTML, estado 200. |
| DELETE /entidades/{id} | Cuerpo vacío, estado 200. |
| GET /estado | Fragmentos OOB de contadores y selectores. |
| GET /salud | Estado 204 si PostgreSQL está disponible. |

Los errores de formulario preservan el DOM y devuelven un mensaje escapado por templ. El HTML inicial, los fragmentos y el cuerpo vacío de borrado usan los mismos repositorios del TP5.

## Verificación

- Pruebas Go con PostgreSQL real y Hurl. Se mantienen los casos de fechas, longitudes, duplicados, referencias y métodos HTTP.
- Chrome real con HTMX descargado del CDN. Cuatro altas, cuatro borrados y una única navegación completa.
- Autor referenciado y usuario repetido muestran el error sin retirar el registro.
- Los contadores, selectores y estados vacíos se actualizan. Un borrador de título y las referencias elegidas se conservan al crear otras entidades.
- Ancho móvil de 390 px sin desborde horizontal.
- El código generado se produce dentro de la imagen. El comando make run-app usa el proyecto bookfly-tp6 y limpia al salir.

La carga remota requiere acceso a https://cdn.jsdelivr.net. No se agregaron archivos JavaScript propios a la entrega.

Referencias oficiales consultadas para respuestas y OOB
- https://htmx.org/docs/
- https://htmx.org/attributes/hx-swap-oob/

Se comprobó en Chrome que el mensaje de borrado bloqueado queda dentro del área visible, junto al registro. El registro se conserva y el mensaje explica que primero se deben eliminar sus libros asociados.

Por decisión del usuario, se exige una edad mínima de 5 años cumplidos al lanzamiento. Go valida la fecha usando el autor consultado y los triggers de PostgreSQL protegen escrituras y cambios de nacimiento. Un bloqueo de la fila del autor coordina las operaciones concurrentes. Para nacidos el 29 de febrero, el quinto aniversario se considera el último día de febrero del año correspondiente, siguiendo el intervalo de PostgreSQL. Las filas anteriores se conservan y la regla se aplica a nuevas escrituras.
