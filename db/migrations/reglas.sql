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

CREATE OR REPLACE FUNCTION bookfly_validar_cronologia_libro() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE nacimiento date;
BEGIN
    -- Se bloquea el autor para coordinar altas con cambios de su nacimiento.
    SELECT fecha_nacimiento INTO nacimiento FROM autor WHERE id_autor = NEW.id_autor FOR UPDATE;
    IF NEW.fecha_lanzamiento < (nacimiento + INTERVAL '5 years')::date THEN
        RAISE EXCEPTION 'El autor debe tener al menos 5 años cumplidos en la fecha de lanzamiento'
            USING ERRCODE = '23514', CONSTRAINT = 'libro_cronologia_autor';
    END IF;
    RETURN NEW;
END
$$;
CREATE OR REPLACE TRIGGER libro_cronologia_autor
BEFORE INSERT OR UPDATE OF id_autor, fecha_lanzamiento ON libro
FOR EACH ROW EXECUTE FUNCTION bookfly_validar_cronologia_libro();

CREATE OR REPLACE FUNCTION bookfly_validar_nacimiento_autor() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF EXISTS (SELECT FROM libro WHERE id_autor = NEW.id_autor AND fecha_lanzamiento < (NEW.fecha_nacimiento + INTERVAL '5 years')::date) THEN
        RAISE EXCEPTION 'El autor debe tener al menos 5 años cumplidos al lanzamiento de todos sus libros'
            USING ERRCODE = '23514', CONSTRAINT = 'autor_cronologia_libros';
    END IF;
    RETURN NEW;
END
$$;
CREATE OR REPLACE TRIGGER autor_cronologia_libros
BEFORE UPDATE OF fecha_nacimiento ON autor
FOR EACH ROW EXECUTE FUNCTION bookfly_validar_nacimiento_autor();
