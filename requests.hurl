POST http://localhost:8080/autores
Content-Type: application/json
{
  "nombre": "Autor Requests Hurl",
  "pais": "Argentina",
  "fecha_nacimiento": "1980-01-15",
  "biografia": "Autor para probar las operaciones CRUD"
}
HTTP/1.1 201
[Captures]
autor_id: jsonpath "$.id"
[Asserts]
jsonpath "$.nombre" == "Autor Requests Hurl"

GET http://localhost:8080/autores
HTTP/1.1 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$[0].nombre" == "Autor Requests Hurl"

GET http://localhost:8080/autores/{{autor_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.id" == {{autor_id}}
jsonpath "$.nombre" == "Autor Requests Hurl"

PUT http://localhost:8080/autores/{{autor_id}}
Content-Type: application/json
{
  "nombre": "Autor Requests Hurl actualizado",
  "pais": "Argentina",
  "fecha_nacimiento": "1980-01-15",
  "biografia": "Biografia actualizada"
}
HTTP/1.1 200
[Asserts]
jsonpath "$.nombre" == "Autor Requests Hurl actualizado"

GET http://localhost:8080/autores/{{autor_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.nombre" == "Autor Requests Hurl actualizado"

POST http://localhost:8080/editoriales
Content-Type: application/json
{
  "nombre": "Editorial Requests Hurl",
  "sede": "Buenos Aires"
}
HTTP/1.1 201
[Captures]
editorial_id: jsonpath "$.id"
[Asserts]
jsonpath "$.nombre" == "Editorial Requests Hurl"

GET http://localhost:8080/editoriales
HTTP/1.1 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$[0].nombre" == "Editorial Requests Hurl"

GET http://localhost:8080/editoriales/{{editorial_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.id" == {{editorial_id}}
jsonpath "$.nombre" == "Editorial Requests Hurl"

PUT http://localhost:8080/editoriales/{{editorial_id}}
Content-Type: application/json
{
  "nombre": "Editorial Requests Hurl actualizada",
  "sede": "La Plata"
}
HTTP/1.1 200
[Asserts]
jsonpath "$.nombre" == "Editorial Requests Hurl actualizada"

GET http://localhost:8080/editoriales/{{editorial_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.sede" == "La Plata"

POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "usuario_requests_hurl",
  "email": "usuario_requests_hurl@example.com",
  "contrasena": "clave123"
}
HTTP/1.1 201
[Captures]
usuario_id: jsonpath "$.id"
[Asserts]
jsonpath "$.nombre" == "usuario_requests_hurl"
jsonpath "$.email" == "usuario_requests_hurl@example.com"

GET http://localhost:8080/usuarios
HTTP/1.1 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$[0].nombre" == "usuario_requests_hurl"

GET http://localhost:8080/usuarios/{{usuario_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.id" == {{usuario_id}}
jsonpath "$.nombre" == "usuario_requests_hurl"

PUT http://localhost:8080/usuarios/{{usuario_id}}
Content-Type: application/json
{
  "nombre": "usuario_requests_hurl_actualizado",
  "email": "usuario_requests_hurl_actualizado@example.com",
  "contrasena": "clave456"
}
HTTP/1.1 200
[Asserts]
jsonpath "$.nombre" == "usuario_requests_hurl_actualizado"

GET http://localhost:8080/usuarios/{{usuario_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.email" == "usuario_requests_hurl_actualizado@example.com"

POST http://localhost:8080/libros
Content-Type: application/json
{
  "titulo": "Libro Requests Hurl",
  "id_autor": {{autor_id}},
  "id_editorial": {{editorial_id}},
  "fecha_lanzamiento": "2020-01-15",
  "cantidad_paginas": 250
}
HTTP/1.1 201
[Captures]
libro_id: jsonpath "$.id"
[Asserts]
jsonpath "$.titulo" == "Libro Requests Hurl"

GET http://localhost:8080/libros
HTTP/1.1 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$[0].titulo" == "Libro Requests Hurl"

GET http://localhost:8080/libros/{{libro_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.id" == {{libro_id}}
jsonpath "$.titulo" == "Libro Requests Hurl"

PUT http://localhost:8080/libros/{{libro_id}}
Content-Type: application/json
{
  "titulo": "Libro Requests Hurl actualizado",
  "id_autor": {{autor_id}},
  "id_editorial": {{editorial_id}},
  "fecha_lanzamiento": "2020-01-15",
  "cantidad_paginas": 300
}
HTTP/1.1 200
[Asserts]
jsonpath "$.cantidad_paginas" == 300

GET http://localhost:8080/libros/{{libro_id}}
HTTP/1.1 200
[Asserts]
jsonpath "$.titulo" == "Libro Requests Hurl actualizado"

# Pruebas de validaciones y errores

POST http://localhost:8080/autores
Content-Type: application/json
```
{"nombre":
```
HTTP/1.1 400
[Asserts]
jsonpath "$.error" == "Cuerpo JSON inválido"

POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "   ",
  "email": "nuevo@example.com",
  "contrasena": "clave123"
}
HTTP/1.1 400

PUT http://localhost:8080/usuarios/{{usuario_id}}
Content-Type: application/json
{
  "nombre": "   ",
  "email": "nuevo@example.com",
  "contrasena": "clave123"
}
HTTP/1.1 400

POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "usuario_nuevo",
  "email": "incorrecto",
  "contrasena": "clave123"
}
HTTP/1.1 400

POST http://localhost:8080/libros
Content-Type: application/json
{
  "titulo": "Libro de prueba",
  "id_autor": {{autor_id}},
  "id_editorial": {{editorial_id}},
  "fecha_lanzamiento": "2020-01-15",
  "cantidad_paginas": 0
}
HTTP/1.1 400

POST http://localhost:8080/libros
Content-Type: application/json
{
  "titulo": "Libro de prueba",
  "id_autor": {{autor_id}},
  "id_editorial": {{editorial_id}},
  "fecha_lanzamiento": "incorrecta",
  "cantidad_paginas": 250
}
HTTP/1.1 400

GET http://localhost:8080/usuarios/abc
HTTP/1.1 400

PATCH http://localhost:8080/usuarios/{{usuario_id}}
HTTP/1.1 405
[Asserts]
header "Allow" == "GET, PUT, DELETE"

POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "usuario_requests_hurl_actualizado",
  "email": "nuevo@example.com",
  "contrasena": "clave123"
}
HTTP/1.1 409

POST http://localhost:8080/libros
Content-Type: application/json
{
  "titulo": "Libro de prueba",
  "id_autor": 2147483647,
  "id_editorial": {{editorial_id}},
  "fecha_lanzamiento": "2020-01-15",
  "cantidad_paginas": 250
}
HTTP/1.1 409

DELETE http://localhost:8080/autores/{{autor_id}}
HTTP/1.1 409

DELETE http://localhost:8080/libros/{{libro_id}}
HTTP/1.1 204

GET http://localhost:8080/libros/{{libro_id}}
HTTP/1.1 404

DELETE http://localhost:8080/usuarios/{{usuario_id}}
HTTP/1.1 204

GET http://localhost:8080/usuarios/{{usuario_id}}
HTTP/1.1 404

DELETE http://localhost:8080/autores/{{autor_id}}
HTTP/1.1 204

GET http://localhost:8080/autores/{{autor_id}}
HTTP/1.1 404

DELETE http://localhost:8080/editoriales/{{editorial_id}}
HTTP/1.1 204

GET http://localhost:8080/editoriales/{{editorial_id}}
HTTP/1.1 404

GET http://localhost:8080/libros
HTTP/1.1 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$" count == 0
