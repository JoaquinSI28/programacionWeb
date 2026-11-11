# TP4 — registro de implementación y verificación

Fecha: 11 de noviembre de 2026.

## Base y alcance

Rama `tp4`, creada desde `tp3` en `4e70149de054ec058f2c6d617e9511cf06da3826`. El siguiente TP deberá crear `tp5` desde el commit final de `tp4`, sin modificar la rama entregada. La rama se publicó en GitHub con autorización del usuario en el commit 9940c80.

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

TP4 implementado y validado. TP5 pendiente: deberá transformar esta interfaz en componentes templ y formularios clásicos, conservando el dominio y PostgreSQL. La versión inicial está publicada en GitHub. Las correcciones de auditoría están pendientes de aprobación del mensaje y de commit.


## Correcciones de la auditoría

- Cada consulta actualiza su sección al llegar, sin esperar las demás; una respuesta tardía de una carga anterior se descarta mediante un contador de carga; los controles se habilitan por sección. Las referencias del formulario de libros esperan a autores/editoriales.
- Un temporizador de 15 segundos cancela peticiones que no terminan. Abarca tanto la recepción de cabeceras como la lectura del cuerpo. El plazo es una decisión del proyecto, no un número fijado por la cátedra.
- Los campos del formulario enviado se bloquean hasta terminar para evitar perder ediciones posteriores. Se liberan en finally, tanto al confirmar como al fallar. Los errores conservan los valores.
- Si un POST supera el plazo, se informa que el resultado no pudo confirmarse y que debe consultarse antes de repetir; cancelar la espera no demuestra que la base no haya guardado el cambio.
- Las editoriales duplicadas continúan permitidas, según lo acordado.

### Relación con el material

08-view.pdf, páginas físicas 4–7: responsabilidades y estado de la interfaz; 29–43: asincronía, Web APIs, temporizadores y promesas; 19–27: DOM, propiedades y eventos. tp4.pdf, cursada páginas 2–3: fetch, formularios y actualización de listas. Se mantienen JavaScript del navegador y la API Go del TP3, sin bibliotecas nuevas ni HTMX/templ.

AbortController es un complemento nativo para cancelar fetch; no se encontró un ejemplo literal de esta API en las diapositivas. Se verificó su comportamiento en la referencia MDN enlazada como recurso en la guía: https://developer.mozilla.org/en-US/docs/Web/API/AbortController/abort .

### Pruebas de regresión agregadas

- Retener la respuesta de usuarios: las otras tres listas deben mostrarse y sus formularios poder usarse. Al superar el plazo, usuarios muestra error y recupera controles.
- Retener la respuesta de un alta confirmada: los campos permanecen bloqueados; al recibirla se limpian y habilitan.
- Hacer dos cargas superpuestas y liberar la respuesta antigua al final: no debe reemplazar datos ni errores de la carga actual.
- Retener un POST sin confirmación: al superar el plazo se conservan los valores, se liberan controles y se advierte que no debe repetirse sin comprobar el resultado.

Validación de las correcciones: compilación y arranque en Docker correctos; suite completa de navegador con Chrome y PostgreSQL reales aprobada, incluidos los cuatro casos de regresión; node --check y git diff --check correctos. El código Go y el esquema SQL no cambiaron.
