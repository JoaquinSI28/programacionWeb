# TP5: HTML del servidor y formularios clásicos con Post/Redirect/Get.
GET http://localhost:8080/
HTTP 200
[Asserts]
header "Content-Type" == "text/html; charset=utf-8"
xpath "count(//script)" == 0
xpath "count(//form[@id='libros-form'])" == 1

POST http://localhost:8080/editoriales
[Form]
nombre: Editorial Hurl TP5
sede: Tandil
HTTP 303
[Asserts]
header "Location" == "/"

POST http://localhost:8080/autores
[Form]
nombre: Autor Hurl TP5
pais: Argentina
fecha_nacimiento: 1980-01-15
biografia: Autor de prueba
HTTP 303
[Asserts]
header "Location" == "/"

POST http://localhost:8080/usuarios
[Form]
nombre: lector_hurl_tp5
email: lector_hurl_tp5@example.com
contrasena: clave_hurl_tp5
HTTP 303
[Asserts]
header "Location" == "/"

GET http://localhost:8080/
HTTP 200
[Captures]
autor_id: xpath "string(//select[@name='id_autor']/option[2]/@value)"
editorial_id: xpath "string(//select[@name='id_editorial']/option[2]/@value)"
autor_delete: xpath "string(//ul[@id='autores-list']//form/@action)"
editorial_delete: xpath "string(//ul[@id='editoriales-list']//form/@action)"
usuario_delete: xpath "string(//ul[@id='usuarios-list']//form/@action)"
[Asserts]
body contains "Editorial Hurl TP5"
body contains "Autor Hurl TP5"
body not contains "clave_hurl_tp5"

POST http://localhost:8080/libros
[Form]
titulo: Libro Hurl TP5
id_autor: {{autor_id}}
id_editorial: {{editorial_id}}
fecha_lanzamiento: 2020-01-15
cantidad_paginas: 250
HTTP 303
[Asserts]
header "Location" == "/"

GET http://localhost:8080/
HTTP 200
[Captures]
libro_delete: xpath "string(//ul[@id='libros-list']//form/@action)"
[Asserts]
xpath "count(//ul[@id='libros-list']/li[h3 or div/h3])" == 1
body contains "Libro Hurl TP5"

# Recargar el GET no repite el alta.
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "count(//ul[@id='libros-list']/li[h3 or div/h3])" == 1

POST http://localhost:8080/usuarios
[Form]
nombre: lector_hurl_tp5
email: lector_hurl_tp5@example.com
contrasena: clave_hurl_tp5
HTTP 409
[Asserts]
body contains "ya está registrado"
body not contains "clave_hurl_tp5"

POST http://localhost:8080/editoriales
[Form]
nombre:
sede: Tandil
HTTP 400
[Asserts]
body contains "El nombre no puede estar vacío"

POST http://localhost:8080{{autor_delete}}
HTTP 409
[Asserts]
body contains "datos relacionados"

GET http://localhost:8080{{libro_delete}}
HTTP 405

POST http://localhost:8080{{libro_delete}}
HTTP 303

POST http://localhost:8080{{autor_delete}}
HTTP 303

POST http://localhost:8080{{editorial_delete}}
HTTP 303

POST http://localhost:8080{{usuario_delete}}
HTTP 303

GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "count(//ul[@id='libros-list']//form)" == 0
xpath "count(//ul[@id='autores-list']//form)" == 0
xpath "count(//ul[@id='editoriales-list']//form)" == 0
xpath "count(//ul[@id='usuarios-list']//form)" == 0

GET http://localhost:8080/app.js
HTTP 404
