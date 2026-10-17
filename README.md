# Programacion Web
Aquí se realizaran entregas de trabajos prácticos para la materia 
Programación Web de la Universidad Nacional de la Provincia de Buenos Aires 
del curso 2026.

# Dominio
Un registro de libros leídos será el tema principal de la app.

# Persistencia

La aplicación utiliza PostgreSQL como sistema de persistencia. El esquema de la base de datos se encuentra en `db/schema/schema.sql` y se monta en el contenedor de PostgreSQL para crear las tablas cuando se inicializa una base nueva.

## Modelo de datos

- `Autor`: almacena el nombre, país, fecha de nacimiento y biografía de cada autor.
- `Editorial`: almacena el nombre y la sede de cada editorial.
- `Usuario`: almacena el nombre de usuario, correo electrónico y contraseña. El nombre de usuario y el correo son únicos.
- `Libro`: almacena el título, fecha de lanzamiento, cantidad de páginas y sus relaciones con un autor y una editorial.
- `Lee`: relaciona usuarios con libros y almacena el estado de lectura y la reseña. La combinación `id_usuario` e `id_libro` es única.

Las relaciones principales son:

```text
Autor 1 ---- N Libro N ---- 1 Editorial
Usuario N ---- N Libro (a través de Lee)
```

Las claves foráneas garantizan que un libro solo pueda referenciar autores y editoriales existentes, y que un registro de lectura solo pueda referenciar usuarios y libros existentes.

## Acceso a los datos

Las consultas SQL se encuentran en `db/queries/`. sqlc las transforma en código Go dentro de `db/sqlc/`, que es una carpeta generada automáticamente y no se versiona. La configuración de esta generación está en `sqlc.yaml`.

Docker Compose utiliza un volumen llamado `pgdata` para conservar los datos de PostgreSQL entre ejecuciones. Los comandos de pruebas eliminan ese volumen antes de comenzar y al finalizar porque los tests utilizan una base de datos temporal.

# Pruebas de persistencia

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


## Ejecución y pruebas del CRUD

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

### Endpoints

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

### Validaciones y errores

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
