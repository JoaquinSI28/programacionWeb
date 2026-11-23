# Guía del catálogo de Bookfly

La página reúne libros, autores, editoriales y usuarios. Podés pasar de una sección a otra usando los enlaces del encabezado. Si todavía no abriste Bookfly, seguí los pasos de [Cómo ejecutar Bookfly](../README.md#cómo-ejecutar-bookfly).

## Recorrido por el catálogo

1. **Cargá una autora.** En **Autores**, ingresá «Ana Ríos», país «Argentina», fecha de nacimiento 12 de abril de 1980 y una biografía breve. Al elegir **Agregar autor**, Ana aparece en el listado y queda disponible en el formulario de libros.
2. **Prepará una edición.** Agregá la editorial «Sur», con sede en Tandil. En **Libros**, escribí el título «El camino del sur», elegí a Ana y a Sur, e ingresá 120 páginas. Antes de guardar, agregá otra editorial, «Norte», con sede en Azul. Al volver al libro, el título y la selección de Sur siguen ahí. Norte también aparece entre las opciones.
3. **Guardá el libro.** Si ingresás el 11 de abril de 1985 como lanzamiento, al enviar el formulario aparece un aviso porque Ana todavía tenía cuatro años. Cambiá solo el día al 12 de abril y volvé a elegir **Agregar libro**. La edición aparece en la lista con su autora, editorial, fecha y cantidad de páginas.
4. **Probá con una editorial repetida.** Volvé a cargar «SUR» con sede en Tandil. El formulario indica que ya existe y Sur sigue apareciendo una sola vez. Si cambiás la sede a Junín, podés guardar esa editorial. La sede permite distinguirlas al elegir una para un libro.
5. **Registrá un usuario.** En **Usuarios**, completá el nombre, un correo válido y una contraseña de prueba. En la lista se muestran el nombre y el correo. Recargá la página para volver a consultar los registros guardados en PostgreSQL.
6. **Eliminá los registros en orden.** Intentá eliminar a Ana mientras el libro siga en el catálogo. Junto al botón aparece el motivo del rechazo y tanto la autora como su libro siguen visibles. Eliminá primero «El camino del sur». Ahora podés eliminar a Ana y a la editorial Sur de Tandil. Las opciones de ambos desaparecen del formulario de libros. El usuario se elimina desde su propia fila.

### Ver qué intercambia la página con el servidor

Abrí la pestaña **Red** de las herramientas de desarrollo y recargá la página una vez. La solicitud `GET /` trae el documento completo. En la página se carga HTMX desde jsDelivr. Si los botones dejan de actualizar las listas, comprobá que el archivo `htmx.min.js` esté disponible. [Plantilla de la página](../views/layout.templ)

Con el registro de red abierto, agregá una editorial y luego eliminala. Abrí cada solicitud y mirá su respuesta. Durante esas dos acciones debe haber una sola carga de tipo **Documento**, la que hiciste al principio.

| Acción | Solicitud | Qué se ve en la respuesta |
| --- | --- | --- |
| Guardar una editorial nueva | `POST /editoriales` | Estado 200 y una lista HTML con todas las editoriales, incluida la nueva. |
| Repetir el nombre y la sede | `POST /editoriales` | Estado 409 y un mensaje que aparece en el formulario. |
| Eliminar una editorial usada por un libro | `DELETE /editoriales/{id}` | Estado 409 y un mensaje junto al botón de esa editorial. |
| Eliminar una editorial sin libros | `DELETE /editoriales/{id}` | Estado 200 con cuerpo vacío. HTMX retira la fila correspondiente. |

Después de un alta vas a ver otra solicitud, por ejemplo `GET /formularios/editoriales`, que trae el formulario vacío. Al cambiar las editoriales también se consulta `GET /opciones/editoriales` para renovar el selector del libro. El título y los demás campos que estabas completando siguen en la página.

## Reglas del catálogo

| Registro | Cuándo se considera repetido |
| --- | --- |
| Autor | Coinciden nombre, país y fecha de nacimiento. |
| Editorial | Coinciden nombre y sede. |
| Libro | Coinciden título, autor, editorial y fecha de lanzamiento. |
| Usuario | Ya está registrado el nombre de usuario o el correo. Cada uno es único por separado. |

Al comparar textos se ignoran las mayúsculas, los espacios de los extremos y los espacios repetidos. Por eso cambiar solo la biografía de un autor o la cantidad de páginas de una edición tampoco alcanza para crear otro registro con la misma identidad.

El catálogo registra libros ya publicados. La fecha de lanzamiento y la de nacimiento admiten hasta el día de hoy en Argentina. El autor debe tener cinco años cumplidos el día del lanzamiento. Si nació el 29 de febrero, se toma el 28 de febrero como quinto cumpleaños. La cantidad de páginas debe ser un entero positivo.

## Cómo se guarda un libro

El formulario envía los datos a `POST /libros`. El servidor interpreta los campos y la capa de lógica comprueba, entre otras cosas, la edad del autor al publicarse el libro. La consulta del nacimiento y el guardado de la edición se hacen en una misma transacción. Así, si alguien cambia esa fecha al mismo tiempo, el libro y el autor quedan sujetos a la misma regla. Los repositorios usan las consultas de sqlc para acceder a PostgreSQL. [Manejo del formulario](../API/web.go) · [Lógica del libro](../logic/libro.go) · [Transacción](../repository/transaccion.go)

Después del alta, el servidor devuelve la lista actualizada como HTML generado con templ. HTMX reemplaza la lista de libros y se pide un formulario vacío para la próxima carga. Si falla una validación, el mensaje aparece en el formulario y se pueden corregir los datos ingresados. [Formularios](../views/entity_form.templ) · [Listas](../views/entity_list.templ)

## Pruebas

Ejecutá estos comandos desde la raíz. Las pruebas de Go y las de Hurl preparan una base limpia al comenzar y eliminan el volumen al terminar. El nombre `bookfly-pruebas` las mantiene separadas del catálogo que guardaste con `bookfly-biblioteca`.

| Comando | Qué comprueba |
| --- | --- |
| `COMPOSE_PROJECT_NAME=bookfly-pruebas make test` | Las reglas de dominio en Go, la persistencia en PostgreSQL, las operaciones simultáneas y las respuestas del servidor web. |
| `COMPOSE_PROJECT_NAME=bookfly-pruebas make test-hurl` | Las rutas HTTP con la aplicación en marcha. Envía formularios, consulta el HTML y comprueba las respuestas a altas, errores y bajas. |
| `COMPOSE_PROJECT_NAME=bookfly-pruebas make test-all` | Ejecuta primero las pruebas de Go y después las de Hurl. |

Para consultar un caso concreto, los límites de edad están en las [pruebas de lógica](../logic/reglas_test.go), los conflictos entre escrituras en las [pruebas de base de datos](../db/reglas_test.go) y los intercambios HTTP en la [secuencia de Hurl](../requests.hurl).
