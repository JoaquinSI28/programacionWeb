# TP6. Recorrido HTTP sobre la base vacía que prepara make test-hurl.
GET http://localhost:8080/
HTTP 200
[Asserts]
header "Content-Type" == "text/html; charset=utf-8"
xpath "count(//head/script[@src and contains(@src,'htmx.org')])" == 1
xpath "count(//form[@hx-swap='outerHTML' and @hx-post=@action and @hx-target=concat('#',substring-after(@action,'/'),'-list')])" == 4

POST http://localhost:8080/editoriales
[Form]
nombre: Editorial Hurl
sede: Tandil
HTTP 200
[Captures]
editorial_delete: xpath "string(//ul[@id='editoriales-list']//button/@hx-delete)"
[Asserts]
xpath "count(//body/*)" == 1
xpath "count(//ul[@id='editoriales-list']//button[@hx-target='closest li' and @hx-swap='outerHTML'])" == 1

POST http://localhost:8080/autores
[Form]
nombre: Autor Hurl
pais: Argentina
fecha_nacimiento: 1980-01-15
biografia: Autor de prueba
HTTP 200
[Captures]
autor_delete: xpath "string(//ul[@id='autores-list']//button/@hx-delete)"
[Asserts]
xpath "count(//body/*)" == 1
xpath "count(//ul[@id='autores-list']//button[@hx-target='closest li' and @hx-swap='outerHTML'])" == 1

POST http://localhost:8080/usuarios
[Form]
nombre: lector_hurl
email: lector_hurl@example.com
contrasena: clave_hurl
HTTP 200
[Captures]
usuario_delete: xpath "string(//ul[@id='usuarios-list']//button/@hx-delete)"
[Asserts]
xpath "count(//body/*)" == 1
xpath "count(//ul[@id='usuarios-list']//button[@hx-target='closest li' and @hx-swap='outerHTML'])" == 1
body not contains "clave_hurl"

GET http://localhost:8080/
HTTP 200
[Captures]
autor_id: xpath "string(//select[@name='id_autor']/option[normalize-space(@value)!='']/@value)"
editorial_id: xpath "string(//select[@name='id_editorial']/option[normalize-space(@value)!='']/@value)"
[Asserts]
body contains "Editorial Hurl"
body contains "Autor Hurl"

# Si el usuario eligió un autor o una editorial, al refrescar opciones se conserva su selección.
GET http://localhost:8080/opciones/autores?id_autor={{autor_id}}
HTTP 200
[Asserts]
xpath "count(//option[@selected and @value='{{autor_id}}'])" == 1

GET http://localhost:8080/opciones/editoriales?id_editorial={{editorial_id}}
HTTP 200
[Asserts]
xpath "count(//option[@selected and @value='{{editorial_id}}'])" == 1

POST http://localhost:8080/libros
[Form]
titulo: Libro Hurl
id_autor: {{autor_id}}
id_editorial: {{editorial_id}}
fecha_lanzamiento: 2020-01-15
cantidad_paginas: 250
HTTP 200
[Captures]
libro_delete: xpath "string(//ul[@id='libros-list']//button/@hx-delete)"
[Asserts]
xpath "count(//body/*)" == 1
xpath "count(//ul[@id='libros-list']//button[@hx-target='closest li' and @hx-swap='outerHTML'])" == 1
body contains "Libro Hurl"

GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "count(//ul[@id='libros-list']//button[@hx-delete])" == 1
body contains "Libro Hurl"

# Volver a consultar la página conserva el libro y no repite el alta.
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "count(//ul[@id='libros-list']//button[@hx-delete])" == 1

POST http://localhost:8080/usuarios
[Form]
nombre: lector_hurl
email: lector_hurl@example.com
contrasena: clave_hurl
HTTP 409
[Asserts]
xpath "count(//*[@id='usuarios-form-error' and @hx-swap-oob])" == 1
body contains "ya está registrado"
body not contains "clave_hurl"

POST http://localhost:8080/editoriales
[Form]
nombre:
sede: Tandil
HTTP 400
[Asserts]
xpath "count(//*[@id='editoriales-form-error' and @hx-swap-oob])" == 1
body contains "El nombre no puede estar vacío"

DELETE http://localhost:8080{{autor_delete}}
HTTP 409
[Asserts]
xpath "count(//*[@id='autores-{{autor_id}}-error' and @hx-swap-oob])" == 1
body contains "tiene libros asociados"

GET http://localhost:8080{{libro_delete}}
HTTP 405

DELETE http://localhost:8080{{libro_delete}}
HTTP 200
[Asserts]
body == ""

DELETE http://localhost:8080{{autor_delete}}
HTTP 200
[Asserts]
body == ""

DELETE http://localhost:8080{{editorial_delete}}
HTTP 200
[Asserts]
body == ""

DELETE http://localhost:8080{{usuario_delete}}
HTTP 200
[Asserts]
body == ""

GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "count(//ul//button[@hx-delete])" == 0
xpath "count(//select/option[normalize-space(@value)!=''])" == 0

# Una selección cuyo registro se borró deja sólo la opción vacía.
GET http://localhost:8080/opciones/autores?id_autor={{autor_id}}
HTTP 200
[Asserts]
xpath "count(//option)" == 1
xpath "count(//option[@value=''])" == 1

GET http://localhost:8080/opciones/editoriales?id_editorial={{editorial_id}}
HTTP 200
[Asserts]
xpath "count(//option)" == 1
xpath "count(//option[@value=''])" == 1

GET http://localhost:8080/app.js
HTTP 404
