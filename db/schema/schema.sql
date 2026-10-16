CREATE TABLE Autor (
    id_autor SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    pais VARCHAR(100) NOT NULL,
    fecha_nacimiento DATE NOT NULL,
    biografia TEXT NOT NULL
);

CREATE TABLE Editorial (
    id_editorial SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    sede VARCHAR(255) NOT NULL
);

CREATE TABLE Usuario (
    id_usuario SERIAL PRIMARY KEY,
    nombre_usuario VARCHAR(100) NOT NULL UNIQUE,
    mail VARCHAR(255) NOT NULL UNIQUE,
    contrasena VARCHAR(255) NOT NULL
);

CREATE TABLE Libro (
    id_libro SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    id_autor INT NOT NULL,
    id_editorial INT NOT NULL,
    fecha_lanzamiento DATE NOT NULL,
    cantidad_paginas INT NOT NULL,
    FOREIGN KEY (id_autor) REFERENCES Autor(id_autor),
    FOREIGN KEY (id_editorial) REFERENCES Editorial(id_editorial)
);

CREATE TABLE Lee (
    id_usuario INT NOT NULL,
    id_libro INT NOT NULL,
    estado VARCHAR(50) NOT NULL,
    resena TEXT NOT NULL,
    PRIMARY KEY (id_usuario, id_libro),
    FOREIGN KEY (id_usuario) REFERENCES Usuario(id_usuario),
    FOREIGN KEY (id_libro) REFERENCES Libro(id_libro)
);
