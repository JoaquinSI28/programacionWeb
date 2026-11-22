# Programación Web · Bookfly

Proyecto de cursada de Programación Web, UNICEN, 2026. La rama `tp6` continúa la biblioteca del TP5. Permite agregar y eliminar libros, autores, editoriales y usuarios desde una página que Go y templ renderizan en el servidor. HTMX actualiza la lista que cambió.

## Ejecutar la entrega del TP6

Se necesita Git, make y Docker iniciado con el comando `docker compose`. La primera construcción requiere conexión para descargar las imágenes y dependencias. El navegador también necesita acceso a jsDelivr para cargar HTMX.

```bash
git clone --branch tp6 --single-branch https://github.com/JoaquinSI28/programacionWeb.git
cd programacionWeb
make run-app
```

1. Esperá a que la terminal muestre **Bookfly disponible en**. El comando prepara una base limpia, genera el código de sqlc y templ, compila Go dentro de Docker y comprueba que PostgreSQL y la aplicación responden.
2. Abrí en Safari o en otro navegador el enlace completo que imprime la terminal. Docker asigna un puerto libre en tu equipo. Mantené la terminal abierta mientras usás la biblioteca.
3. Para terminar, presioná **Ctrl+C** en esa terminal. `make run-app` elimina los contenedores y el volumen del proyecto `bookfly-tp6` tanto al inicio como al salir. Los datos de esa ejecución son temporales.

Para conservar una biblioteca entre sesiones, usá `make up`, consultá la dirección con `docker compose -p bookfly-tp6 port api 8080` y abrila con el prefijo `http://`. Detenela con `make stop`. Volver a usar `make run-app` o las pruebas sobre ese mismo proyecto borra su volumen.

### Comprobar el funcionamiento

Agregá un autor y una editorial. Elegilos en el formulario de libros y guardá un libro ya publicado, con una fecha en la que el autor tuviera al menos cinco años. El libro aparece en la lista sin recargar la página. Intentá borrar el autor mientras el libro lo referencia. El registro debe seguir visible con la explicación junto al botón. Borrá primero el libro y después el autor.

Para ejecutar las pruebas de regresión, cerrá la aplicación con Ctrl+C y corré este comando desde la raíz del proyecto.

```bash
make test-all
```

`make test` ejecuta las pruebas de Go con PostgreSQL. `make test-hurl` recorre la aplicación por HTTP. `make test-all` ejecuta ambas suites. Cada una prepara su base y elimina su volumen al terminar. La guía del TP6 detalla qué observar en Safari en cada punto de la consigna, cómo verificar un alta duplicada y qué parte comprueba cada suite. [Guía de verificación del TP6](docs/TP6.md)

## Dominio y persistencia

Bookfly registra libros, autores, editoriales y usuarios. La relación de lectura entre un usuario y un libro se conserva en la capa de datos.

La aplicación utiliza PostgreSQL como sistema de persistencia. El esquema de la base de datos se encuentra en `db/schema/schema.sql` y se monta en el contenedor de PostgreSQL para crear las tablas cuando se inicializa una base nueva.

### Modelo de datos

- `Autor`: almacena el nombre, país, fecha de nacimiento y biografía de cada autor.
- `Editorial`: almacena el nombre y la sede de cada editorial.
- `Usuario`: almacena el nombre de usuario, correo electrónico y hash de la contraseña. El nombre de usuario y el correo son únicos.
- `Libro`: almacena el título, fecha de lanzamiento, cantidad de páginas y sus relaciones con un autor y una editorial.
- `Lee`: relaciona usuarios con libros y almacena el estado de lectura y la reseña. La combinación `id_usuario` e `id_libro` es única.

Las relaciones principales son:

```text
Autor 1 ---- N Libro N ---- 1 Editorial
Usuario N ---- N Libro (a través de Lee)
```

Las claves foráneas garantizan que un libro solo pueda referenciar autores y editoriales existentes, y que un registro de lectura solo pueda referenciar usuarios y libros existentes.

### Acceso a los datos

Las consultas SQL se encuentran en `db/queries/`. sqlc las transforma en código Go dentro de `db/sqlc/`, que es una carpeta generada automáticamente y no se versiona. La configuración de esta generación está en `sqlc.yaml`.

Docker Compose utiliza un volumen llamado `pgdata` para conservar los datos de PostgreSQL entre ejecuciones. Los comandos de pruebas eliminan ese volumen antes de comenzar y al finalizar porque los tests utilizan una base de datos temporal.

## Etapas anteriores

Los comandos históricos se ejecutan en la rama indicada. El TP3 expone una API JSON, el TP4 incorpora la interfaz estática y el TP5 usa formularios HTML clásicos. Las guías de esas dos interfaces están en sus respectivas ramas. [Guía del TP4](https://github.com/JoaquinSI28/programacionWeb/blob/tp4/docs/TP4.md) · [Guía del TP5](https://github.com/JoaquinSI28/programacionWeb/blob/tp5/docs/TP5.md)

<details>
<summary>TP3 · Persistencia y API JSON</summary>

### Pruebas de persistencia

Solo es necesario clonar el repositorio, ubicarse en su carpeta raíz y ejecutar:

```bash
make test
```

El comando se encarga de:

1. Eliminar los contenedores y el volumen de una ejecución anterior, y levantar PostgreSQL con valores de desarrollo predeterminados.
2. Generar automáticamente el código de `db/sqlc/`.
3. Esperar a que PostgreSQL esté listo.
4. Ejecutar los tests de Go.
5. Detener y eliminar los contenedores al finalizar, incluso si un test falla.

Si se necesitan otros datos de conexión, se puede crear un archivo `.env` en la raíz antes de ejecutar `make test`, usando estas variables:

```env
POSTGRES_DB=prueba
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
```

El archivo `.env` es local y no debe subirse al repositorio.


### Ejecución y pruebas del CRUD

Se necesita Docker iniciado, Docker Compose y make. Para ejecutar las pruebas de persistencia con `make test` tambien se necesita Go 1.26.7 o superior. sqlc y Hurl se ejecutan dentro de Docker.

Para iniciar la aplicación:

```bash
make up
```

Genera los archivos de sqlc, construye la imagen de la API y levanta PostgreSQL y el servidor. Espera a que ambos servicios estén saludables. La aplicación queda disponible en `http://localhost:8080`.

Para probar la API de manera rapida:

```bash
make test-hurl
```

Borra los contenedores y el volumen anterior, prepara PostgreSQL, genera sqlc, construye e inicia la API y ejecuta `requests.hurl`. Las pruebas verifican creación, listado, búsqueda por ID, actualización y eliminación de usuarios, libros, autores y editoriales. Tambien se prueban datos inválidos, registros inexistentes, duplicados y registros relacionados. Hurl se ejecuta dentro de Docker.

**Antes y despues de las pruebas se eliminan los contenedores y el volumen de PostgreSQL, incluidos sus datos.**

Para ejecutar las pruebas de persistencia y de la API:

```bash
make test-all
```

Para ejecutar el mismo archivo directamente, primero iniciar la aplicación con `make up` sobre una base sin datos de pruebas anteriores y tener Hurl instalado:

```bash
hurl --test requests.hurl
```

Para detener la aplicación conservando los datos:

```bash
docker compose down
```

Para borrar los contenedores, imágenes y volúmenes del proyecto, incluidos los datos de PostgreSQL:

```bash
docker compose --profile test down --rmi all --volumes --remove-orphans
```

#### Endpoints

En este TP3 se exponen usuarios, autores, editoriales y libros. La relacion `Lee` queda en la capa de datos.

Las rutas son `/usuarios`, `/autores`, `/editoriales` y `/libros`. Para cada una se pueden realizar estas operaciones:

| Metodo | Ruta | Descripcion | Respuesta |
| --- | --- | --- | --- |
| POST | /entidades | Crear un registro | 201 y el objeto creado |
| GET | /entidades | Listar los registros | 200 y un array JSON |
| GET | /entidades/{id} | Obtener un registro | 200 y el objeto |
| PUT | /entidades/{id} | Actualizar un registro | 200 y el objeto actualizado |
| DELETE | /entidades/{id} | Eliminar un registro | 204 sin contenido |

`entidades` se reemplaza por el nombre de la ruta correspondiente. POST y PUT reciben JSON con `Content-Type: application/json`. En PUT se envian todos los campos.

Ejemplos de los cuerpos para crear o actualizar:

Usuario:
```json
{"nombre":"usuario1","email":"usuario1@example.com","contrasena":"clave123"}
```

Autor:
```json
{"nombre":"Autor de prueba","pais":"Argentina","fecha_nacimiento":"1980-01-15","biografia":"Biografia del autor"}
```

Editorial:
```json
{"nombre":"Editorial de prueba","sede":"Buenos Aires"}
```

Libro (el autor y la editorial deben existir):
```json
{"titulo":"Libro de prueba","id_autor":1,"id_editorial":1,"fecha_lanzamiento":"2020-01-15","cantidad_paginas":250}
```

Las respuestas incluyen los campos enviados y el `id` del registro. En usuarios no se devuelve la contraseña. Si no hay registros, el listado devuelve `[]`.

#### Validaciones y errores

- Los nombres, el titulo, el pais, la biografia y la sede no pueden estar vacíos ni contener solo espacios.
- El email debe tener un formato válido y la contraseña debe tener al menos 3 caracteres.
- Los IDs y la cantidad de paginas deben ser mayores a cero.
- Las fechas se envian con formato `YYYY-MM-DD`. La fecha de lanzamiento no puede ser futura.
- El nombre de usuario y el email no pueden repetirse.
- No se puede guardar un libro con un autor o editorial inexistentes, ni eliminar un registro que tiene datos relacionados.

Los errores se devuelven como JSON, por ejemplo `{"error":"Usuario no encontrado"}`.

| Codigo | Motivo |
| --- | --- |
| 400 | JSON, ID o datos inválidos |
| 404 | No existe el registro solicitado |
| 405 | El metodo no esta permitido para esa ruta |
| 409 | Nombre de usuario o email repetido, o conflicto con registros relacionados |
| 500 | Error interno del servidor |

#### Aclaraciones
En esta entrega las contraseñas se almacenan en texto plano. Se optó por dejar el hashing de contraseñas y la autenticación para mas adelante.

</details>

<details>
<summary>TP5 · Ejecución, formularios y pruebas</summary>

En la rama `tp5`, Bookfly continúa la interfaz del TP4. El servidor Go consulta PostgreSQL mediante sqlc y usa componentes templ para mostrar los libros, autores, editoriales y usuarios en una página HTML. La relación `Lee` sigue en la capa de datos.

Con Git, Docker Compose y make instalados y Docker iniciado, ejecutar desde una terminal:

```bash
git clone --branch tp5 --single-branch https://github.com/JoaquinSI28/programacionWeb.git
cd programacionWeb
make run-app
```

Si ya se clonó la rama, alcanza con ejecutar `make run-app` desde la raíz. El comando limpia los contenedores y el volumen del proyecto `bookfly-tp5`, genera el código de sqlc y templ, compila la aplicación y espera a que PostgreSQL y el servidor estén listos. Go y los generadores se ejecutan dentro de la imagen de Docker. El arranque y las pruebas pueden leer el clon sin escribir archivos generados en él.

Abrir la dirección que aparece junto a **Bookfly disponible en**. Docker elige un puerto local disponible y la base de datos usa la red interna del proyecto. Al salir con **Ctrl+C** se eliminan los contenedores y el volumen del TP5, incluidos los datos de esa ejecución. Para trabajar conservando los datos entre sesiones, usar `make up` y `make stop`. En ese modo, `docker compose -p bookfly-tp5 port api 8080` muestra la dirección que se abre con HTTP.

### Formularios y pruebas del TP5

1. Agregar un autor y una editorial. Ambos aparecen como opciones en el formulario de libros.
2. Crear un libro con esos datos. El formulario envía un POST y el servidor guarda el registro, responde con una redirección 303 y vuelve a mostrar la biblioteca mediante `GET /`.
3. Recargar la página y comprobar que el libro aparece una sola vez. El recorrido completo funciona con JavaScript desactivado.
4. Para probar el borrado, eliminar primero el libro y después el autor y la editorial asociados.

El formulario de usuarios admite contraseñas de 3 a 255 caracteres. En el TP5 se guarda un hash con sal aleatoria y al abrir una base anterior se migran las contraseñas existentes. El listado de usuarios consulta únicamente el ID, el nombre y el correo.

La fecha de nacimiento y la de lanzamiento admiten hasta el día de hoy en Argentina. Para evitar duplicados se compara el nombre, país y fecha de nacimiento del autor; el nombre y sede de la editorial; y el título, autor, editorial y fecha de lanzamiento de cada edición. El nombre de usuario y el correo son únicos por separado. Las comparaciones ignoran mayúsculas y espacios repetidos. El formulario devuelve 400 para datos inválidos y 409 si el registro ya existe.

Desde esta rama, `make test` ejecuta las pruebas de Go con PostgreSQL y `make test-hurl` prueba los formularios y las respuestas HTML. Para ejecutar ambos:

```bash
make test-all
```

Las pruebas crean y eliminan el volumen del proyecto.

</details>
