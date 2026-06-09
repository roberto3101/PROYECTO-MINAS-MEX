-- =============================================================================
-- AGENTE E: TRAZABILIDAD, AUDITORÍA Y APPEND-ONLY
-- Sistema Mina — Batería de invariantes de diseño
-- Ejecutar como: psql -U postgres -h 127.0.0.1 -p 5433 -d mina -f E_trazabilidad.sql
-- =============================================================================

\set ON_ERROR_STOP off
\pset footer off

-- ─────────────────────────────────────────────────────────────────────────────
-- CABECERA
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '=========================================================' AS sep;
SELECT 'AGENTE E — TRAZABILIDAD, AUDITORÍA Y APPEND-ONLY'    AS titulo;
SELECT 'Fecha de ejecucion: ' || now()::text                  AS ejecutado_en;
SELECT '=========================================================' AS sep2;

-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 1: MULTI-TENANT — TODA tabla base de negocio tiene id_empresa
-- Schemas auditados: catalogos, produccion, planeacion, reconciliacion,
--                    beneficio, estandares, costos, inversiones
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";
SELECT 'INVARIANTE 1: MULTI-TENANT — columna id_empresa en TODA tabla base' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";

-- Tablas que NO tienen id_empresa
WITH tablas_negocio AS (
    SELECT table_schema, table_name
    FROM information_schema.tables
    WHERE table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
    AND table_type = 'BASE TABLE'
),
con_id_empresa AS (
    SELECT table_schema, table_name
    FROM information_schema.columns
    WHERE column_name = 'id_empresa'
    AND table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
),
sin_id_empresa AS (
    SELECT t.table_schema, t.table_name
    FROM tablas_negocio t
    LEFT JOIN con_id_empresa c
           ON t.table_schema = c.table_schema
          AND t.table_name   = c.table_name
    WHERE c.table_name IS NULL
)
SELECT
    CASE WHEN COUNT(*) = 0
         THEN 'PASA  — todas las tablas de negocio tienen id_empresa (' ||
              (SELECT COUNT(*) FROM tablas_negocio) || ' tablas revisadas)'
         ELSE 'FALLA — ' || COUNT(*) || ' tabla(s) sin id_empresa (ver detalle abajo)'
    END AS "Resultado INV-1"
FROM sin_id_empresa;

-- Detalle de excepciones (vacío si pasa)
SELECT table_schema || '.' || table_name AS "Tabla sin id_empresa (INV-1)"
FROM (
    SELECT t.table_schema, t.table_name
    FROM information_schema.tables t
    WHERE t.table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
    AND t.table_type = 'BASE TABLE'
    AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns c
        WHERE c.table_schema = t.table_schema
          AND c.table_name   = t.table_name
          AND c.column_name  = 'id_empresa'
    )
) exc
ORDER BY table_schema, table_name;


-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 2: APPEND-ONLY — tablas de evento no deben tener
--   actualizado_en ni eliminado_en
-- Tablas evento: produccion.acarreo_viaje, rezagado_ciclo, barrenacion_avance,
--   demora_equipo, consumo_explosivo; reconciliacion.medicion;
--   beneficio.ley_metalurgica, recuperacion; inversiones.consumo_acero
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";
SELECT 'INVARIANTE 2: APPEND-ONLY — eventos sin actualizado_en ni eliminado_en' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";

WITH eventos AS (
    SELECT unnest(ARRAY[
        'produccion.acarreo_viaje',
        'produccion.rezagado_ciclo',
        'produccion.barrenacion_avance',
        'produccion.demora_equipo',
        'produccion.consumo_explosivo',
        'reconciliacion.medicion',
        'beneficio.ley_metalurgica',
        'beneficio.recuperacion',
        'inversiones.consumo_acero'
    ]) AS tabla_fqn
),
eventos_split AS (
    SELECT
        split_part(tabla_fqn, '.', 1) AS esquema,
        split_part(tabla_fqn, '.', 2) AS tabla
    FROM eventos
),
columnas_prohibidas AS (
    SELECT e.esquema, e.tabla, c.column_name AS col_prohibida
    FROM eventos_split e
    JOIN information_schema.columns c
      ON c.table_schema = e.esquema
     AND c.table_name   = e.tabla
     AND c.column_name  IN ('actualizado_en', 'eliminado_en')
),
tablas_violadoras AS (
    SELECT esquema || '.' || tabla  AS tabla_evento,
           string_agg(col_prohibida, ', ' ORDER BY col_prohibida) AS columnas_extra
    FROM columnas_prohibidas
    GROUP BY esquema, tabla
)
SELECT
    CASE WHEN COUNT(*) = 0
         THEN 'PASA  — ninguna tabla de evento tiene actualizado_en / eliminado_en'
         ELSE 'FALLA — ' || COUNT(*) || ' tabla(s) de evento con columnas mutables (ver detalle)'
    END AS "Resultado INV-2"
FROM tablas_violadoras;

SELECT tabla_evento AS "Tabla evento violadora (INV-2)",
       columnas_extra AS "Columnas prohibidas presentes"
FROM (
    SELECT e.esquema || '.' || e.tabla AS tabla_evento,
           string_agg(c.column_name, ', ' ORDER BY c.column_name) AS columnas_extra
    FROM (
        SELECT unnest(ARRAY['produccion','produccion','produccion','produccion',
                             'produccion','reconciliacion','beneficio','beneficio','inversiones'])
            AS esquema,
               unnest(ARRAY['acarreo_viaje','rezagado_ciclo','barrenacion_avance','demora_equipo',
                             'consumo_explosivo','medicion','ley_metalurgica','recuperacion','consumo_acero'])
            AS tabla
    ) e
    JOIN information_schema.columns c
      ON c.table_schema = e.esquema
     AND c.table_name   = e.tabla
     AND c.column_name  IN ('actualizado_en', 'eliminado_en')
    GROUP BY e.esquema, e.tabla
) v
ORDER BY tabla_evento;


-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 3: TRAZABILIDAD COMPLETA en mutables
-- Toda tabla con eliminado_en DEBE tener TAMBIÉN:
--   actualizado_en, creado_en, creado_por_usuario_id,
--   actualizado_por_usuario_id, eliminado_por_usuario_id
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";
SELECT 'INVARIANTE 3: TRAZABILIDAD COMPLETA en tablas mutables' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";

WITH tablas_con_eliminado AS (
    SELECT DISTINCT table_schema, table_name
    FROM information_schema.columns
    WHERE column_name = 'eliminado_en'
    AND table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
),
columnas_requeridas AS (
    SELECT unnest(ARRAY[
        'actualizado_en',
        'creado_en',
        'creado_por_usuario_id',
        'actualizado_por_usuario_id',
        'eliminado_por_usuario_id'
    ]) AS col_requerida
),
faltantes AS (
    SELECT t.table_schema, t.table_name, r.col_requerida
    FROM tablas_con_eliminado t
    CROSS JOIN columnas_requeridas r
    WHERE NOT EXISTS (
        SELECT 1 FROM information_schema.columns c
        WHERE c.table_schema = t.table_schema
          AND c.table_name   = t.table_name
          AND c.column_name  = r.col_requerida
    )
),
resumen AS (
    SELECT table_schema || '.' || table_name AS tabla,
           string_agg(col_requerida, ', ' ORDER BY col_requerida) AS columnas_faltantes
    FROM faltantes
    GROUP BY table_schema, table_name
)
SELECT
    CASE WHEN COUNT(*) = 0
         THEN 'PASA  — todas las tablas con eliminado_en tienen trazabilidad completa (' ||
              (SELECT COUNT(*) FROM tablas_con_eliminado) || ' tablas mutables revisadas)'
         ELSE 'FALLA — ' || COUNT(*) || ' tabla(s) con trazabilidad incompleta (ver detalle)'
    END AS "Resultado INV-3"
FROM resumen;

SELECT tabla AS "Tabla mutable incompleta (INV-3)",
       columnas_faltantes AS "Columnas faltantes"
FROM (
    SELECT t.table_schema || '.' || t.table_name AS tabla,
           string_agg(r.col_requerida, ', ' ORDER BY r.col_requerida) AS columnas_faltantes
    FROM (
        SELECT DISTINCT table_schema, table_name
        FROM information_schema.columns
        WHERE column_name = 'eliminado_en'
        AND table_schema IN (
            'catalogos','produccion','planeacion','reconciliacion',
            'beneficio','estandares','costos','inversiones'
        )
    ) t
    CROSS JOIN (
        SELECT unnest(ARRAY[
            'actualizado_en','creado_en',
            'creado_por_usuario_id','actualizado_por_usuario_id','eliminado_por_usuario_id'
        ]) AS col_requerida
    ) r
    WHERE NOT EXISTS (
        SELECT 1 FROM information_schema.columns c
        WHERE c.table_schema = t.table_schema
          AND c.table_name   = t.table_name
          AND c.column_name  = r.col_requerida
    )
    GROUP BY t.table_schema, t.table_name
) det
ORDER BY tabla;


-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 4: TIMESTAMPS CON DEFAULT — creado_en y actualizado_en deben
--   tener DEFAULT now() y ser NOT NULL en TODA tabla que las posea
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";
SELECT 'INVARIANTE 4: DEFAULT now() y NOT NULL en timestamps de auditoría' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";

WITH ts_cols AS (
    SELECT table_schema, table_name, column_name, column_default, is_nullable
    FROM information_schema.columns
    WHERE column_name IN ('creado_en', 'actualizado_en')
    AND table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
),
violaciones AS (
    SELECT table_schema || '.' || table_name AS tabla,
           column_name,
           COALESCE(column_default, '(sin default)') AS col_default,
           is_nullable,
           CASE
               WHEN column_default IS NULL THEN 'Sin DEFAULT'
               WHEN column_default NOT ILIKE '%now()%'
                AND column_default NOT ILIKE '%CURRENT_TIMESTAMP%'
                THEN 'DEFAULT incorrecto: ' || column_default
               ELSE NULL
           END AS problema_default,
           CASE WHEN is_nullable = 'YES' THEN 'Permite NULL' ELSE NULL END AS problema_null
    FROM ts_cols
    WHERE column_default IS NULL
       OR (column_default NOT ILIKE '%now()%' AND column_default NOT ILIKE '%CURRENT_TIMESTAMP%')
       OR is_nullable = 'YES'
)
SELECT
    CASE WHEN COUNT(*) = 0
         THEN 'PASA  — creado_en y actualizado_en tienen DEFAULT now() y NOT NULL en todas las tablas'
         ELSE 'FALLA — ' || COUNT(*) || ' columna(s) con DEFAULT o nulabilidad incorrectos (ver detalle)'
    END AS "Resultado INV-4"
FROM violaciones;

SELECT tabla AS "Tabla (INV-4)",
       column_name AS "Columna",
       col_default AS "Default actual",
       is_nullable AS "Nullable",
       COALESCE(problema_default, '') ||
       CASE WHEN problema_default IS NOT NULL AND problema_null IS NOT NULL THEN ' + ' ELSE '' END ||
       COALESCE(problema_null, '') AS "Problema"
FROM (
    SELECT table_schema || '.' || table_name AS tabla,
           column_name,
           COALESCE(column_default, '(sin default)') AS col_default,
           is_nullable,
           CASE
               WHEN column_default IS NULL THEN 'Sin DEFAULT'
               WHEN column_default NOT ILIKE '%now()%'
                AND column_default NOT ILIKE '%CURRENT_TIMESTAMP%'
                THEN 'DEFAULT incorrecto: ' || column_default
               ELSE NULL
           END AS problema_default,
           CASE WHEN is_nullable = 'YES' THEN 'Permite NULL' ELSE NULL END AS problema_null
    FROM information_schema.columns
    WHERE column_name IN ('creado_en', 'actualizado_en')
    AND table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
    AND (
        column_default IS NULL
        OR (column_default NOT ILIKE '%now()%' AND column_default NOT ILIKE '%CURRENT_TIMESTAMP%')
        OR is_nullable = 'YES'
    )
) det
ORDER BY tabla, column_name;


-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 5: ÍNDICES ÚNICOS PARCIALES — los UNIQUE en tablas mutables
--   (con eliminado_en) deben contener WHERE eliminado_en IS NULL
--   para permitir re-alta tras baja lógica
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";
SELECT 'INVARIANTE 5: UNIQUE parciales (WHERE eliminado_en IS NULL) en mutables' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";

WITH tablas_con_eliminado AS (
    SELECT DISTINCT table_schema AS esquema, table_name AS tabla
    FROM information_schema.columns
    WHERE column_name = 'eliminado_en'
    AND table_schema IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
),
unique_no_parciales AS (
    SELECT i.schemaname, i.tablename, i.indexname, i.indexdef
    FROM pg_indexes i
    JOIN tablas_con_eliminado t
      ON i.schemaname = t.esquema
     AND i.tablename  = t.tabla
    WHERE i.indexdef ILIKE '%UNIQUE%'
    -- Excluir PKs (son por id, no hace falta que sean parciales)
    AND i.indexname NOT LIKE '%_pkey'
    -- El índice NO tiene cláusula WHERE
    AND i.indexdef NOT ILIKE '%WHERE%'
)
SELECT
    CASE WHEN COUNT(*) = 0
         THEN 'PASA  — todos los índices UNIQUE de negocio en tablas mutables son parciales'
         ELSE 'FALLA — ' || COUNT(*) || ' índice(s) UNIQUE no parciales en tablas con eliminado_en (ver detalle)'
    END AS "Resultado INV-5"
FROM unique_no_parciales;

SELECT schemaname || '.' || tablename AS "Tabla (INV-5)",
       indexname AS "Índice UNIQUE no parcial",
       indexdef   AS "Definición"
FROM (
    SELECT i.schemaname, i.tablename, i.indexname, i.indexdef
    FROM pg_indexes i
    WHERE i.schemaname IN (
        'catalogos','produccion','planeacion','reconciliacion',
        'beneficio','estandares','costos','inversiones'
    )
    AND i.indexdef ILIKE '%UNIQUE%'
    AND i.indexname NOT LIKE '%_pkey'
    AND i.indexdef NOT ILIKE '%WHERE%'
    AND EXISTS (
        SELECT 1 FROM information_schema.columns c
        WHERE c.table_schema = i.schemaname
          AND c.table_name   = i.tablename
          AND c.column_name  = 'eliminado_en'
    )
) det
ORDER BY schemaname, tablename, indexname;


-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 5b: UNIQUE no parciales en tablas APPEND-ONLY (sin eliminado_en)
-- Estas tablas no necesitan WHERE eliminado_en IS NULL (son inmutables),
-- pero un UNIQUE total en eventos puede impedir reingreso de datos corregidos.
-- Se reporta como ADVERTENCIA para revisión del equipo, no como FALLA técnica.
-- ─────────────────────────────────────────────────────────────────────────────
SELECT 'INVARIANTE 5b: UNIQUE no parciales en tablas append-only (advertencia)' AS "INV-5b";

SELECT i.schemaname || '.' || i.tablename AS "Tabla append-only (INV-5b)",
       i.indexname  AS "Indice UNIQUE total",
       i.indexdef   AS "Definicion"
FROM pg_indexes i
WHERE i.schemaname IN (
    'catalogos','produccion','planeacion','reconciliacion',
    'beneficio','estandares','costos','inversiones'
)
AND i.indexdef ILIKE '%UNIQUE%'
AND i.indexname NOT LIKE '%_pkey'
AND i.indexdef NOT ILIKE '%WHERE%'
AND NOT EXISTS (
    SELECT 1 FROM information_schema.columns c
    WHERE c.table_schema = i.schemaname
      AND c.table_name   = i.tablename
      AND c.column_name  = 'eliminado_en'
)
ORDER BY i.schemaname, i.tablename, i.indexname;


-- ─────────────────────────────────────────────────────────────────────────────
-- INVARIANTE 6: CICLO FUNCIONAL DE BAJA LÓGICA Y RE-ALTA (BEGIN…ROLLBACK)
-- Secuencia sobre catalogos.obra:
--   a) Dar de baja lógica una obra existente (UPDATE eliminado_en = now())
--   b) Verificar que la vista activa ya no la muestra
--   c) Insertar OTRA obra con el MISMO codigo (debe permitirse por índice parcial)
--   d) Verificar que la nueva se ve como activa
--   e) ROLLBACK total
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";
SELECT 'INVARIANTE 6: CICLO FUNCIONAL — baja lógica + re-alta (BEGIN...ROLLBACK)' AS "";
SELECT '─────────────────────────────────────────────────────────' AS "";

BEGIN;

-- Captura del registro original
DO $$
DECLARE
    v_id               uuid;
    v_id_empresa       uuid;
    v_id_mina          uuid;
    v_codigo           text;
    v_nuevo_id         uuid;
    v_visible_tras_baja  int;
    v_visible_nueva    int;
    v_resultado        text;
BEGIN
    -- 6a) Tomar la primera obra activa del seed
    SELECT id, id_empresa, id_mina, codigo
    INTO   v_id, v_id_empresa, v_id_mina, v_codigo
    FROM   catalogos.obra
    WHERE  eliminado_en IS NULL
    ORDER  BY creado_en
    LIMIT  1;

    IF v_id IS NULL THEN
        RAISE NOTICE 'INV-6: OMITIDO — no hay obras activas en la BD (seed no cargado)';
        RETURN;
    END IF;

    RAISE NOTICE 'INV-6: Obra seleccionada id=% codigo=% empresa=% mina=%',
        v_id, v_codigo, v_id_empresa, v_id_mina;

    -- 6b) Dar de baja lógica
    UPDATE catalogos.obra
    SET    eliminado_en             = now(),
           eliminado_por_usuario_id = '22222222-2222-2222-2222-222222222222'
    WHERE  id = v_id;

    -- 6c) Verificar que ya NO aparece con filtro eliminado_en IS NULL
    SELECT COUNT(*) INTO v_visible_tras_baja
    FROM   catalogos.obra
    WHERE  id           = v_id
    AND    eliminado_en IS NULL;

    IF v_visible_tras_baja = 0 THEN
        RAISE NOTICE 'INV-6-PASO-b: CORRECTO — obra dada de baja no visible para activos';
    ELSE
        RAISE NOTICE 'INV-6-PASO-b: ERROR — obra dada de baja sigue visible en activos (%)!', v_visible_tras_baja;
    END IF;

    -- 6d) Re-alta: insertar con mismo codigo; el índice parcial debe permitirlo
    BEGIN
        INSERT INTO catalogos.obra (
            id_empresa, id_mina, codigo, nombre,
            creado_por_usuario_id, actualizado_por_usuario_id
        )
        VALUES (
            v_id_empresa, v_id_mina, v_codigo,
            'RE-ALTA TEST — debe ser rollback',
            '22222222-2222-2222-2222-222222222222',
            '22222222-2222-2222-2222-222222222222'
        )
        RETURNING id INTO v_nuevo_id;

        RAISE NOTICE 'INV-6-PASO-c: CORRECTO — re-alta con mismo codigo permitida (nuevo id=%).',
            v_nuevo_id;

        -- 6e) Verificar que la nueva aparece como activa
        SELECT COUNT(*) INTO v_visible_nueva
        FROM   catalogos.obra
        WHERE  id           = v_nuevo_id
        AND    eliminado_en IS NULL;

        IF v_visible_nueva = 1 THEN
            RAISE NOTICE 'INV-6-PASO-d: CORRECTO — nueva obra visible como activa';
        ELSE
            RAISE NOTICE 'INV-6-PASO-d: ERROR — nueva obra NO visible como activa (%)', v_visible_nueva;
        END IF;

    EXCEPTION WHEN unique_violation THEN
        RAISE NOTICE 'INV-6-PASO-c: FALLA — índice UNIQUE no parcial bloqueó la re-alta (unique_violation)';
    END;

END;
$$;

-- Confirmar resultado final del ciclo con conteo real (dentro de la tx)
SELECT
    CASE
        WHEN baja.cnt = 0 AND nueva.cnt = 1
             THEN 'PASA  — ciclo completo: baja lógica oculta el registro; re-alta con mismo codigo permitida y visible'
        WHEN baja.cnt > 0
             THEN 'FALLA — la baja lógica no oculta el registro activo (cnt=' || baja.cnt || ')'
        WHEN nueva.cnt != 1
             THEN 'FALLA — la re-alta no es visible como activo o el INSERT falló (cnt=' || nueva.cnt || ')'
        ELSE      'ESTADO DESCONOCIDO'
    END AS "Resultado INV-6"
FROM
    -- obras con codigo REB-2400W dadas de baja (should = 0 en activos)
    (SELECT COUNT(*) AS cnt FROM catalogos.obra
     WHERE codigo = 'REB-2400W' AND id_empresa = '11111111-1111-1111-1111-111111111111'
     AND eliminado_en IS NULL
     AND id = 'f6660000-0000-0000-0000-000000000001') baja,
    -- nueva re-alta activa con mismo codigo (should = 1)
    (SELECT COUNT(*) AS cnt FROM catalogos.obra
     WHERE codigo = 'REB-2400W' AND id_empresa = '11111111-1111-1111-1111-111111111111'
     AND eliminado_en IS NULL
     AND nombre = 'RE-ALTA TEST — debe ser rollback') nueva;

ROLLBACK;

SELECT 'INV-6: ROLLBACK ejecutado — estado de la BD restaurado sin cambios.' AS "Post-rollback";


-- ─────────────────────────────────────────────────────────────────────────────
-- RESUMEN EJECUTIVO
-- ─────────────────────────────────────────────────────────────────────────────
SELECT '' AS "";
SELECT '=========================================================' AS "";
SELECT 'RESUMEN — Agente E: Trazabilidad, Auditoría y Append-Only' AS "";
SELECT '=========================================================' AS "";

SELECT
    inv,
    descripcion
FROM (VALUES
    ('INV-1', 'Multi-tenant: id_empresa en toda tabla base de negocio'),
    ('INV-2', 'Append-only: eventos sin actualizado_en ni eliminado_en'),
    ('INV-3', 'Trazabilidad completa: mutables con los 5 campos de auditoría'),
    ('INV-4', 'DEFAULT now() + NOT NULL en creado_en / actualizado_en'),
    ('INV-5', 'UNIQUE parciales (WHERE eliminado_en IS NULL) en mutables'),
    ('INV-6', 'Ciclo funcional baja lógica + re-alta (BEGIN...ROLLBACK)')
) AS t(inv, descripcion)
ORDER BY inv;

SELECT 'Fin de la batería E_trazabilidad.sql — revisar resultados arriba.' AS "";
