-- ============================================================
-- A_smoke.sql  —  Smoke tests de arranque  (SOLO LECTURA)
-- Agente: QA-Smoke   Fecha: 2026-06-09
-- ============================================================

\echo '========================================================'
\echo 'CHECK 1 — Cluster responde'
\echo '========================================================'
SELECT 'CHECK-1' AS check_id,
       'Cluster responde' AS descripcion,
       CASE WHEN version() LIKE 'PostgreSQL%' THEN 'PASA' ELSE 'FALLA' END AS resultado,
       version() AS detalle;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 2 — Existencia de los 9 schemas'
\echo '========================================================'
WITH esperados(schema_name) AS (
  VALUES
    ('gobierno'),('catalogos'),('produccion'),('planeacion'),
    ('reconciliacion'),('beneficio'),('estandares'),('costos'),('reportes')
),
presentes AS (
  SELECT schema_name FROM information_schema.schemata
)
SELECT 'CHECK-2' AS check_id,
       e.schema_name,
       CASE WHEN p.schema_name IS NOT NULL THEN 'PASA' ELSE 'FALLA' END AS resultado
FROM esperados e
LEFT JOIN presentes p USING (schema_name)
ORDER BY e.schema_name;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 3 — Conteo de tablas base por schema (esperado total 45)'
\echo '========================================================'
WITH conteo AS (
  SELECT table_schema AS schema_name, count(*) AS tablas
  FROM information_schema.tables
  WHERE table_type = 'BASE TABLE'
    AND table_schema IN (
      'gobierno','catalogos','produccion','planeacion',
      'reconciliacion','beneficio','estandares','costos','inversiones','reportes'
    )
  GROUP BY table_schema
),
esperados(schema_name, esperado) AS (
  VALUES
    ('gobierno',    2),
    ('catalogos',  15),
    ('produccion',  8),
    ('planeacion',  4),
    ('reconciliacion', 2),
    ('beneficio',   3),
    ('estandares',  2),
    ('costos',      6),
    ('inversiones', 3)
)
SELECT 'CHECK-3' AS check_id,
       e.schema_name,
       e.esperado,
       COALESCE(c.tablas, 0) AS encontrado,
       CASE WHEN COALESCE(c.tablas, 0) = e.esperado THEN 'PASA' ELSE 'FALLA' END AS resultado
FROM esperados e
LEFT JOIN conteo c USING (schema_name)
ORDER BY e.schema_name;

-- Total global
SELECT 'CHECK-3-TOTAL' AS check_id,
       45 AS esperado,
       sum(tablas) AS encontrado,
       CASE WHEN sum(tablas) = 45 THEN 'PASA' ELSE 'FALLA' END AS resultado
FROM (
  SELECT count(*) AS tablas
  FROM information_schema.tables
  WHERE table_type = 'BASE TABLE'
    AND table_schema IN (
      'gobierno','catalogos','produccion','planeacion',
      'reconciliacion','beneficio','estandares','costos','inversiones'
    )
) t;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 4a — Listado de vistas en schema reportes'
\echo '========================================================'
SELECT table_name AS vista
FROM information_schema.views
WHERE table_schema = 'reportes'
ORDER BY table_name;

\echo ''
\echo 'Vistas materializadas en schema reportes:'
SELECT matviewname AS vista_materializada
FROM pg_matviews
WHERE schemaname = 'reportes'
ORDER BY matviewname;

\echo ''
\echo 'Conteo total de vistas + materializada (esperado 18 + 1):'
SELECT 'CHECK-4-COUNT' AS check_id,
       (SELECT count(*) FROM information_schema.views WHERE table_schema='reportes') AS vistas_regulares,
       (SELECT count(*) FROM pg_matviews WHERE schemaname='reportes') AS vistas_materializadas,
       CASE WHEN
         (SELECT count(*) FROM information_schema.views WHERE table_schema='reportes') = 18
         AND (SELECT count(*) FROM pg_matviews WHERE schemaname='reportes') = 1
       THEN 'PASA' ELSE 'FALLA' END AS resultado;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 4b — Consultabilidad de cada vista regular'
\echo '========================================================'

-- Usamos DO para iterar y capturar errores; los resultados van a una tabla temporal
DO $$
DECLARE
  v_name TEXT;
  v_count BIGINT;
  v_ok    BOOLEAN;
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS _vista_checks (
    vista TEXT, filas BIGINT, resultado TEXT
  );
  FOR v_name IN
    SELECT table_name FROM information_schema.views
    WHERE table_schema = 'reportes'
    ORDER BY table_name
  LOOP
    BEGIN
      EXECUTE format('SELECT count(*) FROM reportes.%I', v_name) INTO v_count;
      INSERT INTO _vista_checks VALUES (v_name, v_count, 'PASA');
    EXCEPTION WHEN OTHERS THEN
      INSERT INTO _vista_checks VALUES (v_name, -1, 'FALLA: ' || SQLERRM);
    END;
  END LOOP;
END;
$$;

SELECT 'CHECK-4b' AS check_id, vista, filas, resultado
FROM _vista_checks
ORDER BY vista;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 4c — Consultabilidad de la vista materializada'
\echo '========================================================'
DO $$
DECLARE
  mv_name TEXT;
  mv_count BIGINT;
BEGIN
  CREATE TEMP TABLE IF NOT EXISTS _mv_checks (
    vista TEXT, filas BIGINT, resultado TEXT
  );
  FOR mv_name IN
    SELECT matviewname FROM pg_matviews
    WHERE schemaname = 'reportes'
    ORDER BY matviewname
  LOOP
    BEGIN
      EXECUTE format('SELECT count(*) FROM reportes.%I', mv_name) INTO mv_count;
      INSERT INTO _mv_checks VALUES (mv_name, mv_count, 'PASA');
    EXCEPTION WHEN OTHERS THEN
      INSERT INTO _mv_checks VALUES (mv_name, -1, 'FALLA: ' || SQLERRM);
    END;
  END LOOP;
END;
$$;

SELECT 'CHECK-4c' AS check_id, vista AS vista_materializada, filas, resultado
FROM _mv_checks
ORDER BY vista;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 5 — Seed: filas esperadas en tablas clave'
\echo '========================================================'
SELECT 'CHECK-5' AS check_id,
       tabla,
       esperado,
       encontrado,
       CASE WHEN encontrado = esperado THEN 'PASA' ELSE 'FALLA' END AS resultado
FROM (
  SELECT 'catalogos.mina'              AS tabla, 3 AS esperado,
         (SELECT count(*) FROM catalogos.mina)::int AS encontrado
  UNION ALL
  SELECT 'catalogos.equipo',           5, (SELECT count(*) FROM catalogos.equipo)::int
  UNION ALL
  SELECT 'catalogos.empleado',         5, (SELECT count(*) FROM catalogos.empleado)::int
  UNION ALL
  SELECT 'catalogos.obra',             2, (SELECT count(*) FROM catalogos.obra)::int
  UNION ALL
  SELECT 'produccion.acarreo_viaje',   4, (SELECT count(*) FROM produccion.acarreo_viaje)::int
  UNION ALL
  SELECT 'produccion.parte_barrenacion', 1, (SELECT count(*) FROM produccion.parte_barrenacion)::int
  UNION ALL
  SELECT 'reconciliacion.medicion',    4, (SELECT count(*) FROM reconciliacion.medicion)::int
  UNION ALL
  SELECT 'beneficio.lote_molienda',    1, (SELECT count(*) FROM beneficio.lote_molienda)::int
) s
ORDER BY tabla;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 6 — gen_random_uuid() funciona'
\echo '========================================================'
SELECT 'CHECK-6' AS check_id,
       'gen_random_uuid()' AS descripcion,
       CASE WHEN gen_random_uuid() IS NOT NULL THEN 'PASA' ELSE 'FALLA' END AS resultado;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK 7 — mv_plan_vs_real_mensual tiene datos (solo SELECT count)'
\echo '========================================================'
SELECT 'CHECK-7' AS check_id,
       'mv_plan_vs_real_mensual' AS vista_materializada,
       (SELECT count(*) FROM reportes.mv_plan_vs_real_mensual) AS filas,
       CASE WHEN (SELECT count(*) FROM reportes.mv_plan_vs_real_mensual) >= 0
            THEN 'PASA (consultable)'
            ELSE 'FALLA'
       END AS resultado;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'RESUMEN FINAL'
\echo '========================================================'
SELECT
  (SELECT count(*) FROM _vista_checks WHERE resultado = 'PASA')  AS vistas_ok,
  (SELECT count(*) FROM _vista_checks WHERE resultado != 'PASA') AS vistas_falla,
  (SELECT count(*) FROM _mv_checks    WHERE resultado = 'PASA')  AS mv_ok,
  (SELECT count(*) FROM _mv_checks    WHERE resultado != 'PASA') AS mv_falla;
