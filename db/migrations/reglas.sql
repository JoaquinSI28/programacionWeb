-- La comparación conserva el texto que se muestra y normaliza su clave de búsqueda.
CREATE OR REPLACE FUNCTION bookfly_clave(text) RETURNS text
LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE AS $$
    SELECT lower(btrim(regexp_replace($1, '[[:space:]]+', ' ', 'g')))
$$;

CREATE UNIQUE INDEX IF NOT EXISTS autor_identidad ON autor
    (bookfly_clave(nombre), bookfly_clave(pais), fecha_nacimiento);
CREATE UNIQUE INDEX IF NOT EXISTS editorial_identidad ON editorial
    (bookfly_clave(nombre), bookfly_clave(sede));
CREATE UNIQUE INDEX IF NOT EXISTS libro_edicion ON libro
    (bookfly_clave(titulo), id_autor, id_editorial, fecha_lanzamiento);
CREATE UNIQUE INDEX IF NOT EXISTS usuario_nombre_normalizado ON usuario (bookfly_clave(nombre_usuario));
CREATE UNIQUE INDEX IF NOT EXISTS usuario_correo_normalizado ON usuario (bookfly_clave(mail));

-- Al actualizar un volumen anterior se conservan sus filas. Estas restricciones
-- se aplican a las nuevas escrituras aunque ya haya fechas históricas inválidas.
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_constraint WHERE conrelid = 'autor'::regclass AND conname = 'autor_fecha_valida') THEN
        ALTER TABLE autor ADD CONSTRAINT autor_fecha_valida CHECK
            (fecha_nacimiento BETWEEN DATE '0001-01-01' AND (CURRENT_TIMESTAMP AT TIME ZONE 'America/Argentina/Buenos_Aires')::date) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_constraint WHERE conrelid = 'libro'::regclass AND conname = 'libro_fecha_valida') THEN
        ALTER TABLE libro ADD CONSTRAINT libro_fecha_valida CHECK
            (fecha_lanzamiento BETWEEN DATE '0001-01-01' AND (CURRENT_TIMESTAMP AT TIME ZONE 'America/Argentina/Buenos_Aires')::date) NOT VALID;
    END IF;
END
$$;
