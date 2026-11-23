# TP6 · Interfaces reactivas con HTMX

Bookfly parte de la interfaz templ de la rama `tp5`. El navegador recibe la biblioteca completa en el primer `GET /`. Al agregar un registro, actualiza la lista de esa entidad. Al borrarlo, retira su fila. Estas operaciones están disponibles para libros, autores, editoriales y usuarios. Para iniciar la entrega desde un clon nuevo, seguí el [README principal](../README.md#ejecutar-la-entrega-del-tp6).

## Los cuatro puntos de cursada

La consigna es *Trabajo de Cursada: Interfaces Reactivas con HTMX*, páginas 2 y 3 de `tp6.pdf`. La clase de presentación dinámica muestra la instalación de HTMX 4.0.0, los atributos y las respuestas parciales en las diapositivas 38–45 y 54–58 de `10-view-dynamic.pdf`. [Material de la cátedra](https://programacion-web-slides.netlify.app/10-view-dynamic/)

| Punto | Dónde se implementa | Qué comprobar |
| --- | --- | --- |
| 1. Integrar HTMX | El `<head>` de [layout.templ](../views/layout.templ) carga HTMX 4.0.0 desde jsDelivr con verificación de integridad. | En el inspector de red aparece `htmx.min.js` al abrir la biblioteca. |
| 2. Crear sin recarga | Los cuatro formularios de [entity_form.templ](../views/entity_form.templ) envían `hx-post` a su ruta, apuntan al ID de su lista con `hx-target` y usan `hx-swap="outerHTML"`. Tras guardar, el [handler POST](../API/web.go) consulta y renderiza la lista actualizada. | El POST responde 200 con un único `<ul>` como raíz. Después de dos altas distintas aparecen ambos registros una sola vez. |
| 3. Borrar sin recarga | El botón de [entity_list.templ](../views/entity_list.templ) usa `hx-delete`, `hx-target="closest li"` y `hx-swap="outerHTML"`. El [handler DELETE](../API/web.go) toma el ID de la ruta y elimina el registro mediante el repositorio y sqlc. | Un borrado permitido responde 200 con el cuerpo vacío. Desaparece únicamente la fila elegida. |
| 4. Verificar el ciclo en el navegador | [requests.hurl](../requests.hurl) prueba las respuestas HTTP. El recorrido siguiente comprueba el efecto de esas respuestas sobre la página. | Durante las altas y bajas aparecen peticiones parciales en la pestaña Red (Fetch o XHR). En ese intervalo hay una sola carga de tipo documento, la del inicio. |

## Recorrido en Safari

Con `make run-app` en ejecución, abrí la dirección que indica la terminal. Para observar la red, usá **Desarrollo → Mostrar inspector web** y elegí **Red**. Si el menú está oculto, activá las funciones para desarrolladores en **Safari → Configuración → Avanzado**. Con el inspector abierto, recargá una vez para registrar el documento inicial.

1. **Comprobá dos altas consecutivas.** Creá la editorial «Sur», sede «Tandil». En el formulario del libro, elegí Sur y escribí «Primer libro». Ahora creá la editorial «Norte», sede «Azul». Ambas editoriales deben aparecer una sola vez y el borrador debe conservar el título y la selección Sur. Abrí la segunda petición `POST /editoriales` en Red. La respuesta debe contener la lista completa, incluidas las dos editoriales. Las peticiones auxiliares que actualizan el formulario o el selector aparecen por separado.
2. **Creá un libro a partir de opciones nuevas.** Agregá una autora con nombre «Ana», país «Argentina», nacimiento `1980-04-12` y una biografía breve. En el formulario de libro elegí Ana y Sur, escribí «Primer libro», 120 páginas, y probá el lanzamiento `1985-04-11`. La edad mínima se cumple al día siguiente. Debés ver un 400 y una explicación en el formulario, con los datos aún cargados. Cambiá únicamente la fecha a `1985-04-12`. El libro debe aparecer con el autor y la editorial elegidos.
3. **Probá un duplicado.** Volvé a intentar el alta de «SUR», sede «Tandil». Esperá un 409, el aviso dentro del formulario y una sola editorial Sur en la lista. Para comprobar que el error desaparece al corregir los datos, cambiá la sede a «Junín», guardá y eliminá esa nueva editorial. Probá también un usuario con un correo válido y una contraseña de prueba. Deben mostrarse su nombre y su correo.
4. **Intentá borrar un registro relacionado.** En la fila de Ana, pulsá **Eliminar**. La respuesta debe ser 409. El mensaje junto al botón explica que hay libros asociados y conserva a la autora. El libro debe seguir visible.
5. **Comprobá la persistencia.** Antes de recargar, revisá que las acciones anteriores solo agregaron solicitudes parciales. Recargá manualmente. Ana, las dos editoriales, el usuario y el libro deben volver a aparecer desde PostgreSQL, una sola vez cada uno.
6. **Borrá en orden.** Eliminá el libro, luego la autora, la editorial Sur y el usuario. Cada `DELETE /<entidad>/<id>` permitido debe responder 200 con cuerpo vacío. Norte sigue en la lista de editoriales y se puede elegir en el formulario de libros. Al borrar Norte, su opción también desaparece. Estas bajas deben ocurrir sin otra carga de documento.

## Cómo está organizado el proyecto

- **Presentación.** Go usa `net/http` para las rutas y componentes templ para la página, las listas y los formularios. Después de un alta, HTMX pide el formulario vacío de esa entidad. Cuando cambian los autores o editoriales pide las opciones del selector correspondiente. Los demás campos del borrador permanecen en la página y una selección sigue vigente mientras exista su registro. Los errores de alta se muestran en el formulario, y los de baja junto al registro que los originó. [Vistas](../views) · [Handlers](../API/web.go)
- **Lógica y persistencia.** El handler interpreta el formulario y delega las reglas del dominio en `logic`. Los casos de uso que cruzan entidades (`CreateLibro`, `UpdateLibro`, `UpdateAutor`) aplican la regla de edad dentro de `logic` coordinando los repositorios a través del contrato `UnitOfWork` (`logic/uow.go`). Los repositorios (`repository`) escriben y consultan PostgreSQL mediante sqlc y proveen la implementación transaccional `SQLUnitOfWork`. Las claves foráneas y los índices únicos protegen las relaciones e identidades del catálogo. La composición de componentes y el acceso con sqlc se explican en `09-view-static.pdf` (diapositivas 13–21 y 26–33); la separación entre handler, lógica y repositorio en `07-logic.pdf` (diapositivas 11–13); y el patrón Unit of Work y pruebas con mocks en `12-diseño.pdf` (diapositivas 31–32 y 37–40). [Lógica](../logic) · [Repositorios](../repository) · [Consultas](../db/queries)
- **Herramientas.** Los dos módulos externos de Go declarados en `go.mod` son templ y el controlador PostgreSQL `lib/pq`. El navegador descarga HTMX desde el CDN. Los archivos que componen la interfaz son fuentes templ y CSS. La construcción genera el código de sqlc y templ dentro de la imagen. [go.mod](../go.mod) · [Dockerfile](../Dockerfile)

Después de un alta, el evento `load` de la nueva lista pide el formulario vacío. Los selectores del libro escuchan `htmx:after:request` en la sección de autores o editoriales y vuelven a pedir sus opciones con el ID elegido. Escuchar en esas secciones con los modificadores `from` y `target` es una decisión de Bookfly para preservar el borrador. El navegador obtiene tanto el formulario como las opciones en HTML. [Listas](../views/entity_list.templ) · [Selectores](../views/entity_form.templ)

### Rutas que aparecen en el inspector

| Petición | Qué recibe el navegador |
| --- | --- |
| `GET /` | El documento completo con las cuatro secciones. |
| `POST /<entidad>` | La lista completa de esa entidad en HTML y el estado 200. |
| `DELETE /<entidad>/{id}` | Un 200 sin cuerpo. |
| `GET /formularios/<entidad>` | El formulario vacío que se muestra después de un alta exitosa. |
| `GET /opciones/autores` y `GET /opciones/editoriales` | Las opciones del selector correspondiente. El ID elegido se envía como parámetro para mantener la selección si existe. |
| `GET /salud` | Un 204 cuando el servidor puede consultar PostgreSQL. Docker lo usa durante el arranque. |

`<entidad>` representa `libros`, `autores`, `editoriales` o `usuarios`. Un dato inválido devuelve 400, un registro ausente devuelve 404 y un duplicado o conflicto de referencias devuelve 409. En esos casos el servidor devuelve un aviso con `hx-swap-oob` y el identificador del formulario o la fila. HTMX coloca el aviso allí y conserva el registro y los datos que el usuario estaba escribiendo. La clase de presentación dinámica ilustra ese intercambio en las diapositivas 48–49 (páginas 60–61 del PDF) de `10-view-dynamic.pdf`. [Rutas y respuestas](../API/web.go)

### Reglas de la biblioteca

El TP6 pide altas y bajas parciales. Bookfly conserva los criterios de identidad del catálogo y define una edad mínima de publicación como regla propia.

| Una operación entra en conflicto cuando… | Criterio |
| --- | --- |
| Se repite un autor | Mismo nombre, país y nacimiento. |
| Se repite una editorial | Mismo nombre y sede. Dos editoriales del mismo nombre pueden tener sedes distintas. |
| Se repite una edición | Mismo título, autor, editorial y fecha de lanzamiento. |
| Se repite un usuario | Coincide el nombre de usuario o el correo. Cada campo es único por separado. |
| Se borra un registro referenciado | Primero hay que eliminar los registros que lo utilizan. |

Para reconocer duplicados, se ignoran mayúsculas, espacios de los extremos y espacios repetidos. PostgreSQL comprueba la unicidad al escribir, incluso si dos clientes envían altas equivalentes a la vez, y `repository/errors.go` traduce esos errores de base de datos a errores de dominio (`03-data.pdf`, diapositivas 4 y 14). El nacimiento y el lanzamiento admiten hasta el día de hoy en Argentina. Una publicación es válida desde el quinto cumpleaños del autor. Para el 29 de febrero, el quinto aniversario se toma el 28 de febrero. El servidor valida los campos obligatorios y las fechas en `logic`. Al guardar un libro o modificar el nacimiento de un autor, `logic` consulta los datos relacionados y ejecuta `ValidarEdadAlLanzamiento` dentro de un `UnitOfWork`, sin acoplarse a `database/sql`. En producción, `repository.SQLUnitOfWork` ejecuta esa unidad dentro de una transacción `SERIALIZABLE` de PostgreSQL y reintenta si dos operaciones simultáneas entran en conflicto (`40001`). En pruebas unitarias (`logic/reglas_test.go`), un mock de `UnitOfWork` permite verificar la regla sin levantar base de datos.

Las consultas de actualización y borrado usan `RETURNING` con `:one` de sqlc. Así el repositorio puede avisar si el ID ya fue eliminado. La clase de mappers describe esta forma de obtener una fila en la diapositiva 11 de `04-mappersORM.pdf`.

Al iniciar sobre un volumen conservado, la aplicación verifica que puede crear los índices de unicidad. Si los datos anteriores contienen duplicados, el arranque informa el conflicto y conserva los registros. Para reutilizar ese volumen hay que resolver el duplicado antes de volver a iniciar. [Reglas de PostgreSQL](../db/migrations/reglas.sql)

## Pruebas de regresión

| Comando desde la raíz | Qué verifica |
| --- | --- |
| `make test` | Ejecuta las pruebas Go unitarias (con mocks en `logic`) y de integración con PostgreSQL. Incluye la regla de edad mínima aislada en `logic`, altas simultáneas equivalentes, bajas simultáneas y conflictos entre un lanzamiento y un cambio de nacimiento. [Pruebas unitarias de lógica](../logic/reglas_test.go) · [Pruebas de datos](../db/db_test.go) · [Concurrencia y migración](../db/reglas_test.go) · [Pruebas web](../API/web_test.go) |
| `make test-hurl` | Levanta la aplicación y envía peticiones HTTP reales para verificar contenido HTML, estados, altas y borrados. [Secuencia Hurl](../requests.hurl) |
| `make test-all` | Ejecuta ambas suites en orden y devuelve error si alguna falla. [Makefile](../Makefile) |

Las pruebas y `make run-app` eliminan los contenedores y el volumen del proyecto que usan. Si ya tenés datos guardados con `make up`, podés aislar las pruebas con `COMPOSE_PROJECT_NAME=bookfly-tp6-pruebas make test-all`. El recorrido en Safari comprueba además que HTMX cargó desde el CDN y que el navegador aplicó cada respuesta en el lugar esperado.

### Resultado de la verificación del 23 de noviembre de 2026

Se refactorizó la regla de edad mínima al lanzamiento hacia `logic` usando el patrón `UnitOfWork` (`logic/uow.go` y `repository/transaccion.go`), dejando a `repository/autor.go` y `repository/libro.go` como adaptadores puros de persistencia y sumando pruebas unitarias de dominio con mocks.

| Prueba ejecutada | Resultado observado |
| --- | --- |
| `go test ./logic -v` | Pasaron los 6 casos unitarios de `TestReglaEdadMinimaEnLogicConMocks` sin PostgreSQL ni servidor HTTP. |
| `make test-all` | Pasaron todas las pruebas Go contra PostgreSQL 15 (incluyendo concurrencia `SERIALIZABLE` en `db/reglas_test.go` y `API/web_test.go`) y las 22 solicitudes de `requests.hurl`. |

### Resultado de la verificación del 22 de noviembre de 2026

La prueba se ejecutó sobre un clon de la rama `tp6` en `5d88d8f` con los cambios de esta revisión, incluyendo HTMX 4.0.0 y la validación de edad en Go. Se usaron bases vacías y nombres de proyecto Docker exclusivos para la aplicación y las pruebas.

| Prueba ejecutada | Resultado observado |
| --- | --- |
| Arranque y cierre con `make run-app` | La terminal imprimió el enlace local. `GET /` respondió 200 y `GET /salud` respondió 204. Después de Ctrl+C quedaron cero contenedores, volúmenes y redes del proyecto de prueba. |
| `make test-all` | Pasaron las pruebas Go contra PostgreSQL 15 y las 22 solicitudes Hurl. Se comprobó también la limpieza posterior. |
| Recorrido en Chrome con HTMX del CDN | Se crearon y borraron las cuatro entidades con una sola carga del documento en la pestaña original. Una segunda pestaña leyó los registros guardados. Se comprobaron los avisos de error, el límite del quinto cumpleaños, el borrador de libro, las opciones tras altas y bajas y las listas vacías. |
| Doce altas equivalentes de una editorial enviadas a la vez | Una respondió 200 y once respondieron 409. Una consulta directa a PostgreSQL confirmó una sola fila guardada. |
