# Bookfly

Bookfly es una aplicación web para organizar un catálogo de libros. Cada libro reúne su título, autor, editorial, fecha de lanzamiento y cantidad de páginas. Desde el navegador se pueden consultar las listas, agregar registros y eliminarlos. La biblioteca también permite registrar usuarios.

El proyecto se desarrolló en **Programación Web de la Universidad Nacional del Centro de la Provincia de Buenos Aires (UNICEN), cursada 2026**. Esta guía vive en `main`. El código de la versión que describe está en `tp6`.

## Ejecutar Bookfly

Necesitás **Git**, **make** y **Docker** en ejecución con el comando `docker compose` disponible. Docker se encarga de compilar Go, generar las consultas y las vistas, e iniciar PostgreSQL. La primera construcción descarga imágenes y dependencias. El navegador necesita conexión a jsDelivr para cargar HTMX.

### Usar una biblioteca que conserve los datos

```bash
git clone --branch tp6 https://github.com/JoaquinSI28/programacionWeb.git
cd programacionWeb
make up
docker compose -p bookfly-tp6 port api 8080
```

Si ya clonaste el repositorio en `main`, ejecutá `git switch tp6` desde la raíz del clon y continuá con `make up`.

Docker elige un puerto libre. Si el último comando muestra `127.0.0.1:49160`, por ejemplo, abrí `http://127.0.0.1:49160` en tu navegador. Para terminar, ejecutá `make stop`. La próxima vez, `make up` recupera el catálogo guardado en el volumen de PostgreSQL. Volvé a consultar la dirección porque el puerto puede cambiar.

### Probar con una biblioteca temporal

Desde la carpeta del proyecto, podés iniciar una sesión de prueba independiente de tu catálogo habitual.

```bash
COMPOSE_PROJECT_NAME=bookfly-demo make run-app
```

Abrí el enlace que imprime la terminal junto a **Bookfly disponible en**. Al presionar **Ctrl+C** se eliminan los contenedores y el volumen de `bookfly-demo`. Este comando también vacía ese proyecto Docker antes de arrancar.

## Cargar el primer libro

1. En **Autores**, agregá a *Ana Ruiz*, de Argentina, nacida el 12 de abril de 1980. Completá la biografía y guardá el registro.
2. En **Editoriales**, agregá *Editorial Sur* con sede en Tandil. La autora y la editorial ya aparecen en los selectores del formulario de libros.
3. En **Tus libros**, escribí *El jardín de invierno*, elegí a Ana y Editorial Sur, indicá 180 páginas y el 15 de mayo de 2008 como fecha de lanzamiento. Al guardar, el libro aparece en la lista. Si recargás la página, sigue allí.
4. Para quitarlo, usá **Eliminar** en la fila del libro. Si primero intentás eliminar a Ana o a Editorial Sur, Bookfly conserva el registro y avisa junto al botón que primero hay que eliminar los libros asociados.

En **Usuarios** podés registrar un nombre de usuario, correo y contraseña. El listado muestra el nombre y el correo. PostgreSQL guarda la contraseña como un hash con sal.

## Cómo se organiza el catálogo

| Registro | Cuándo se considera duplicado |
| --- | --- |
| Autor | Coinciden el nombre, el país y la fecha de nacimiento. |
| Editorial | Coinciden el nombre y la sede. Un mismo nombre puede usarse en otra sede. |
| Libro | Coinciden el título, el autor, la editorial y la fecha de lanzamiento. |
| Usuario | Ya existe el nombre de usuario o el correo electrónico. |

Para comparar textos se ignoran las diferencias de mayúsculas y los espacios sobrantes. Las fechas de nacimiento y lanzamiento se admiten hasta el día de hoy en Argentina. Al publicarse un libro, su autor debe tener al menos cinco años cumplidos.

En la base de datos, la relación `Lee` vincula a un usuario con un libro y guarda su estado de lectura y reseña. La interfaz actual gestiona las cuatro entidades de la tabla. [Esquema de PostgreSQL](https://github.com/JoaquinSI28/programacionWeb/blob/tp6/db/schema/schema.sql)

## Cómo funciona por dentro

La primera visita carga una página HTML completa. Al agregar un libro, HTMX envía el formulario al servidor Go. El servidor valida los datos, guarda el libro en PostgreSQL, consulta el catálogo actualizado y devuelve la lista en HTML. Al eliminar un registro, el servidor responde con un cuerpo vacío y HTMX retira la fila correspondiente. Esto permite seguir trabajando en la misma página.

| Parte del código | Responsabilidad |
| --- | --- |
| [API](https://github.com/JoaquinSI28/programacionWeb/tree/tp6/API) | Atiende las peticiones HTTP, lee los formularios y sirve los estilos. |
| [Vistas](https://github.com/JoaquinSI28/programacionWeb/tree/tp6/views) | Los componentes templ construyen la página, los formularios y las listas que actualiza HTMX. |
| [Lógica](https://github.com/JoaquinSI28/programacionWeb/tree/tp6/logic) | Define las entidades y valida reglas del catálogo, como la edad del autor al publicar. |
| [Repositorios](https://github.com/JoaquinSI28/programacionWeb/tree/tp6/repository) y [base de datos](https://github.com/JoaquinSI28/programacionWeb/tree/tp6/db) | Los repositorios usan las consultas que genera sqlc. PostgreSQL conserva los registros y protege sus relaciones. |

<details>
<summary>Referencia de rutas HTTP</summary>

En estas rutas, `<entidad>` puede ser `libros`, `autores`, `editoriales` o `usuarios`.

| Petición | Respuesta |
| --- | --- |
| `GET /` | Página completa con los cuatro formularios y sus listas. |
| `POST /<entidad>` | Recibe los campos del formulario. Al guardar, devuelve `200` con la lista actualizada en HTML. |
| `DELETE /<entidad>/<id>` | Devuelve `200` con el cuerpo vacío para que el navegador retire esa fila. |
| `GET /salud` | Devuelve `204` cuando el servidor puede consultar PostgreSQL. |

Los campos inválidos reciben `400`, un ID inexistente recibe `404` al intentar borrarlo y un duplicado o un borrado bloqueado por otra entidad recibe `409`. En el navegador, el mensaje aparece en el formulario o junto a la fila que originó la operación.

</details>

## Ejecutar las pruebas

Desde la raíz del código, usá un proyecto Docker separado para que las pruebas puedan crear y eliminar su propia base.

```bash
COMPOSE_PROJECT_NAME=bookfly-pruebas make test-all
```

El comando ejecuta primero las pruebas de Go con PostgreSQL y después las solicitudes HTTP de Hurl. Se comprueban, entre otros casos, las altas simultáneas de una misma editorial, la edad del autor en la fecha de lanzamiento y el mensaje al intentar borrar un autor con libros asociados. Para ejecutar una sola parte, sustituí `test-all` por `test` o `test-hurl`. [Comandos del proyecto](https://github.com/JoaquinSI28/programacionWeb/blob/tp6/Makefile) · [Pruebas HTTP](https://github.com/JoaquinSI28/programacionWeb/blob/tp6/requests.hurl)

## Historia del proyecto

Cada rama conserva una etapa de la cursada y continúa el trabajo de la anterior.

| Rama | Qué incorporó |
| --- | --- |
| [tp1](https://github.com/JoaquinSI28/programacionWeb/tree/tp1) | Primer servidor HTTP en Go y presentación del dominio. |
| [tp2](https://github.com/JoaquinSI28/programacionWeb/tree/tp2) | Modelo relacional en PostgreSQL, consultas con sqlc y pruebas de persistencia. |
| [tp3](https://github.com/JoaquinSI28/programacionWeb/tree/tp3) | Desarrollo de la API REST del catálogo, con validaciones y pruebas HTTP. |
| [tp4](https://github.com/JoaquinSI28/programacionWeb/tree/tp4) | Interfaz en HTML y CSS que consulta la API desde el navegador con JavaScript. |
| [tp5](https://github.com/JoaquinSI28/programacionWeb/tree/tp5) | Renderizado en el servidor con templ y formularios HTML. |
| [tp6](https://github.com/JoaquinSI28/programacionWeb/tree/tp6) | Altas y bajas con HTMX y actualización de listas sin recargar la página. |
