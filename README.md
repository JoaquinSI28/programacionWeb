# Programacion Web
Aquí se realizaran entregas de trabajos prácticos para la materia 
Programación Web de la Universidad Nacional de la Provincia de Buenos Aires 
del curso 2026.

# Dominio
Un registro de libros leídos será el tema principal de la app.

# Información que se guardara
Cada libro contendrá, id del libro, titulo, id autor, id editorial, fecha de lanzamiento, cantidad de paginas.

Lee (relación Usuario lee libro): Id usuario, Id libro, estado, reseña.

Cada usuario tendrá su nombre de usuario, id usuario, lista leídos (Lee), mail, contraseña.

Cada autor tendrá su nombre, id autor, país, fecha de nacimiento, biografía.

Cada editorial tendrá su nombre, id editorial, sede.

# Instrucciones de uso

Solo es necesario clonar el repositorio, ubicarse en su carpeta raíz y ejecutar:

```bash
make test
```

El comando se encarga de:

1. Levantar PostgreSQL con valores de desarrollo predeterminados.
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

