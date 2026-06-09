-- ============================================================
-- D · Auditoría de Rendimiento e Indexación
-- Agente: DBA Senior
-- Fecha: 2026-06-09
-- SOLO LECTURA: SELECT, EXPLAIN (sin ANALYZE en DML)
-- ============================================================

\echo '=================================================================='
\echo 'BLOQUE 1: FKs SIN ÍNDICE DE RESPALDO'
\echo '=================================================================='

-- 1A. Lista completa de FKs con estado de cobertura por índice
-- Lógica: una FK está "cubierta" si existe un índice cuya primera columna
-- coincide con la columna de la FK (índice simple o compuesto leading).
WITH fks AS (
  SELECT
    tc.constraint_name,
    tc.table_schema,
    tc.table_name,
    kcu.column_name,
    ccu.table_schema  AS ref_schema,
    ccu.table_name    AS ref_table,
    ccu.column_name   AS ref_column
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu
    ON kcu.constraint_name = tc.constraint_name
   AND kcu.table_schema    = tc.table_schema
  JOIN information_schema.referential_constraints rc
    ON rc.constraint_name    = tc.constraint_name
   AND rc.constraint_schema  = tc.table_schema
  JOIN information_schema.key_column_usage ccu
    ON ccu.constraint_name = rc.unique_constraint_name
   AND ccu.table_schema    = rc.unique_constraint_schema
  WHERE tc.constraint_type = 'FOREIGN KEY'
),
idx_cols AS (
  -- primera columna de cada índice (leading column)
  SELECT
    schemaname,
    tablename,
    indexname,
    (regexp_split_to_array(
       regexp_replace(indexdef, '.* USING \w+ \(', '', 'i'),
       '[,)]'
     ))[1] AS leading_col
  FROM pg_indexes
  WHERE schemaname NOT IN ('pg_catalog','information_schema')
)
SELECT
  fks.table_schema                            AS esquema,
  fks.table_name                              AS tabla,
  fks.column_name                             AS columna_fk,
  fks.ref_schema || '.' || fks.ref_table      AS referencia,
  fks.constraint_name,
  CASE WHEN ic.leading_col IS NOT NULL
       THEN 'SI  (' || ic.indexname || ')'
       ELSE '*** FALTA ***'
  END AS indice_respaldo
FROM fks
LEFT JOIN idx_cols ic
  ON  ic.schemaname = fks.table_schema
  AND ic.tablename  = fks.table_name
  AND trim(ic.leading_col) = fks.column_name
ORDER BY
  (ic.leading_col IS NULL) DESC,  -- primero las que faltan
  fks.table_schema,
  fks.table_name,
  fks.column_name;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 1B: RESUMEN - Solo FKs SIN índice de respaldo'
\echo '=================================================================='

WITH fks AS (
  SELECT
    tc.constraint_name,
    tc.table_schema,
    tc.table_name,
    kcu.column_name,
    ccu.table_schema  AS ref_schema,
    ccu.table_name    AS ref_table
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu
    ON kcu.constraint_name = tc.constraint_name
   AND kcu.table_schema    = tc.table_schema
  JOIN information_schema.referential_constraints rc
    ON rc.constraint_name    = tc.constraint_name
   AND rc.constraint_schema  = tc.table_schema
  JOIN information_schema.key_column_usage ccu
    ON ccu.constraint_name = rc.unique_constraint_name
   AND ccu.table_schema    = rc.unique_constraint_schema
  WHERE tc.constraint_type = 'FOREIGN KEY'
),
idx_cols AS (
  SELECT
    schemaname,
    tablename,
    (regexp_split_to_array(
       regexp_replace(indexdef, '.* USING \w+ \(', '', 'i'),
       '[,)]'
     ))[1] AS leading_col
  FROM pg_indexes
  WHERE schemaname NOT IN ('pg_catalog','information_schema')
)
SELECT
  fks.table_schema   AS esquema,
  fks.table_name     AS tabla,
  fks.column_name    AS columna_fk_sin_indice,
  fks.ref_schema || '.' || fks.ref_table AS referencia,
  fks.constraint_name
FROM fks
LEFT JOIN idx_cols ic
  ON  ic.schemaname = fks.table_schema
  AND ic.tablename  = fks.table_name
  AND trim(ic.leading_col) = fks.column_name
WHERE ic.leading_col IS NULL
ORDER BY fks.table_schema, fks.table_name, fks.column_name;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 2: ÍNDICES EXISTENTES POR TABLA'
\echo '=================================================================='

SELECT
  schemaname   AS esquema,
  tablename    AS tabla,
  indexname    AS indice,
  indexdef     AS definicion
FROM pg_indexes
WHERE schemaname NOT IN ('pg_catalog','information_schema')
ORDER BY schemaname, tablename, indexname;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 3A: EXPLAIN de mv_plan_vs_real_mensual (query subyacente)'
\echo '=================================================================='

EXPLAIN
WITH cte_plan AS (
  SELECT p.id_empresa, p.id_mina, mp.periodo_etiqueta AS periodo, sum(mp.toneladas_plan) AS ton_plan
  FROM planeacion.meta_periodo mp
  JOIN planeacion.bloque_programado bp ON bp.id = mp.id_bloque_programado
  JOIN planeacion.plan p ON p.id = bp.id_plan
  WHERE mp.horizonte = 'MENSUAL' AND bp.eliminado_en IS NULL AND mp.eliminado_en IS NULL
  GROUP BY p.id_empresa, p.id_mina, mp.periodo_etiqueta
),
cte_real AS (
  SELECT pa.id_empresa, pa.id_mina, to_char(pa.fecha,'YYYY-MM') AS periodo, sum(av.toneladas) AS ton_real
  FROM produccion.parte_acarreo pa
  JOIN produccion.acarreo_viaje av ON av.id_parte_acarreo = pa.id
  JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
  WHERE tm.es_mineral AND pa.eliminado_en IS NULL
  GROUP BY pa.id_empresa, pa.id_mina, to_char(pa.fecha,'YYYY-MM')
)
SELECT coalesce(pl.id_empresa, re.id_empresa) AS id_empresa,
       coalesce(pl.id_mina, re.id_mina)       AS id_mina,
       coalesce(pl.periodo, re.periodo)       AS periodo,
       coalesce(pl.ton_plan, 0)               AS ton_plan,
       coalesce(re.ton_real, 0)               AS ton_real,
       coalesce(re.ton_real,0) - coalesce(pl.ton_plan,0) AS diferencia,
       CASE WHEN coalesce(pl.ton_plan,0) = 0 THEN NULL
            ELSE round(coalesce(re.ton_real,0) / pl.ton_plan * 100, 2) END AS apego_pct
FROM cte_plan pl
FULL OUTER JOIN cte_real re
  ON pl.id_empresa = re.id_empresa AND pl.id_mina = re.id_mina AND pl.periodo = re.periodo;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 3B: EXPLAIN de v_disponibilidad_equipo (query subyacente)'
\echo '=================================================================='

EXPLAIN
WITH horas AS (
  SELECT id_empresa, id_mina, id_equipo, fecha,
         sum(horometro_final - horometro_inicial) AS horas_operativas
  FROM produccion.parte_acarreo  WHERE eliminado_en IS NULL
  GROUP BY id_empresa, id_mina, id_equipo, fecha
  UNION ALL
  SELECT id_empresa, id_mina, id_equipo, fecha,
         sum(horometro_final - horometro_inicial)
  FROM produccion.parte_rezagado WHERE eliminado_en IS NULL
  GROUP BY id_empresa, id_mina, id_equipo, fecha
),
horas_agg AS (
  SELECT id_empresa, id_mina, id_equipo, fecha, sum(horas_operativas) AS horas_operativas
  FROM horas GROUP BY id_empresa, id_mina, id_equipo, fecha
),
dem AS (
  SELECT id_empresa, id_equipo, fecha, sum(minutos)/60.0 AS horas_demora
  FROM produccion.demora_equipo GROUP BY id_empresa, id_equipo, fecha
)
SELECT h.id_empresa, h.id_mina, e.codigo AS equipo, h.id_equipo, h.fecha,
       round(h.horas_operativas, 2) AS horas_operativas,
       round(coalesce(d.horas_demora, 0), 2) AS horas_demora,
       round(h.horas_operativas / NULLIF(h.horas_operativas + coalesce(d.horas_demora,0), 0) * 100, 2) AS utilizacion_pct
FROM horas_agg h
JOIN catalogos.equipo e ON e.id = h.id_equipo
LEFT JOIN dem d ON d.id_equipo = h.id_equipo AND d.fecha = h.fecha;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 3C: EXPLAIN de v_tiempos_vs_estandar (con LATERAL)'
\echo '=================================================================='

EXPLAIN
SELECT d.id_empresa, d.id_mina, d.equipo, te.codigo AS tipo_equipo, d.fecha,
       d.horas_operativas, d.horas_demora, d.utilizacion_pct,
       et.utilizacion_meta_pct,
       round(d.utilizacion_pct - et.utilizacion_meta_pct, 2) AS desvio_utilizacion_pct
FROM reportes.v_disponibilidad_equipo d
JOIN catalogos.equipo e ON e.id = d.id_equipo
JOIN catalogos.tipo_equipo te ON te.id = e.id_tipo_equipo
LEFT JOIN LATERAL (
  SELECT utilizacion_meta_pct FROM estandares.estandar_tiempo s
  WHERE s.id_empresa = d.id_empresa AND s.id_tipo_equipo = e.id_tipo_equipo
    AND s.eliminado_en IS NULL
  ORDER BY s.vigente_desde DESC LIMIT 1
) et ON true;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 3D: EXPLAIN de v_balance_metalurgico'
\echo '=================================================================='

EXPLAIN
SELECT lm.id_empresa, lm.id_mina, m.nombre AS mina, lm.codigo_lote, lm.fecha, lm.toneladas_molidas,
       max(l.ley_au) FILTER (WHERE l.punto='CABEZA')      AS au_cabeza,
       max(l.ley_au) FILTER (WHERE l.punto='CONCENTRADO') AS au_concentrado,
       max(l.ley_au) FILTER (WHERE l.punto='COLA')        AS au_cola,
       max(l.ley_ag) FILTER (WHERE l.punto='CABEZA')      AS ag_cabeza,
       max(r.recuperacion_pct) FILTER (WHERE r.metal='AU') AS rec_au_pct,
       max(r.recuperacion_pct) FILTER (WHERE r.metal='AG') AS rec_ag_pct
FROM beneficio.lote_molienda lm
JOIN catalogos.mina m ON m.id = lm.id_mina
LEFT JOIN beneficio.ley_metalurgica l ON l.id_lote_molienda = lm.id
LEFT JOIN beneficio.recuperacion r ON r.id_lote_molienda = lm.id
WHERE lm.eliminado_en IS NULL
GROUP BY lm.id_empresa, lm.id_mina, m.nombre, lm.codigo_lote, lm.fecha, lm.toneladas_molidas;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 3E: EXPLAIN de v_obras_prioritarias (subconsultas correlacionadas)'
\echo '=================================================================='

EXPLAIN
SELECT o.id_empresa, o.id_mina, m.nombre AS mina, o.codigo AS obra, o.nombre,
       coalesce((SELECT sum(av.toneladas)
                 FROM produccion.parte_acarreo pa
                 JOIN produccion.acarreo_viaje av ON av.id_parte_acarreo = pa.id
                 JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
                 WHERE pa.id_obra = o.id AND tm.es_mineral AND pa.eliminado_en IS NULL), 0) AS ton_mineral,
       coalesce((SELECT sum(ba.longitud)
                 FROM produccion.parte_barrenacion pb
                 JOIN produccion.barrenacion_avance ba ON ba.id_parte_barrenacion = pb.id
                 WHERE pb.id_obra = o.id AND pb.eliminado_en IS NULL), 0) AS metros_barrenados
FROM catalogos.obra o
JOIN catalogos.mina m ON m.id = o.id_mina
WHERE o.es_prioritaria AND o.eliminado_en IS NULL;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 4: CARDINALIDAD Y TAMAÑO DE TABLAS (para contextualizar EXPLAIN)'
\echo '=================================================================='

SELECT
  n.nspname                                     AS esquema,
  c.relname                                     AS tabla,
  c.reltuples::bigint                           AS filas_estimadas,
  pg_size_pretty(pg_total_relation_size(c.oid)) AS tamano_total,
  pg_size_pretty(pg_indexes_size(c.oid))        AS tamano_indices
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r'
  AND n.nspname NOT IN ('pg_catalog','information_schema','pg_toast')
ORDER BY pg_total_relation_size(c.oid) DESC;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 5: ÍNDICES PARCIALES WHERE eliminado_en IS NULL - COBERTURA'
\echo '=================================================================='

-- ¿Qué tablas tienen eliminado_en pero NO tienen índice parcial sobre sus
-- columnas de filtro principales (id_empresa, id_mina, fecha)?
SELECT
  n.nspname AS esquema,
  c.relname AS tabla,
  -- ¿tiene columna eliminado_en?
  EXISTS (
    SELECT 1 FROM pg_attribute a
    WHERE a.attrelid = c.oid AND a.attname = 'eliminado_en' AND a.attnum > 0
  ) AS tiene_eliminado_en,
  -- ¿tiene al menos un índice parcial WHERE eliminado_en IS NULL?
  EXISTS (
    SELECT 1 FROM pg_indexes ix
    WHERE ix.schemaname = n.nspname
      AND ix.tablename  = c.relname
      AND ix.indexdef   ILIKE '%eliminado_en is null%'
  ) AS tiene_indice_parcial
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r'
  AND n.nspname NOT IN ('pg_catalog','information_schema','pg_toast')
ORDER BY n.nspname, c.relname;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 6: PATRONES DE CONSULTA — índices de reporte faltantes'
\echo '=================================================================='

-- 6A: acarreo_viaje — filtros por tipo_mineral (para balance mineral/tepetate)
-- A escala, la vista v_balance_cargas y mv_plan_vs_real filtrará por id_tipo_mineral
-- usando el predicado tm.es_mineral. Con millones de viajes, este join es el cuello.
\echo '6A: EXPLAIN acarreo_viaje por id_tipo_mineral (patron de reporte frecuente)'
EXPLAIN
SELECT av.id_parte_acarreo, av.toneladas
FROM produccion.acarreo_viaje av
JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
WHERE tm.es_mineral = true;


-- 6B: meta_periodo — filtros por horizonte+periodo (el más usado en reportes)
\echo '6B: EXPLAIN meta_periodo por horizonte + periodo_etiqueta'
EXPLAIN
SELECT mp.id_bloque_programado, sum(mp.toneladas_plan)
FROM planeacion.meta_periodo mp
WHERE mp.horizonte = 'MENSUAL'
  AND mp.periodo_etiqueta = '2026-01'
  AND mp.eliminado_en IS NULL
GROUP BY mp.id_bloque_programado;


-- 6C: demora_equipo — filtros por equipo+fecha (patron v_disponibilidad_equipo)
\echo '6C: EXPLAIN demora_equipo por equipo+fecha (ya tiene ix_demora_equipo_fecha?)'
EXPLAIN
SELECT id_equipo, sum(minutos)
FROM produccion.demora_equipo
WHERE id_empresa = (SELECT id FROM gobierno.empresa LIMIT 1)
  AND fecha BETWEEN '2026-01-01' AND '2026-01-31'
GROUP BY id_equipo;


-- 6D: estandar_tiempo — el LATERAL busca por (id_empresa, id_tipo_equipo, vigente_desde DESC)
\echo '6D: EXPLAIN estandar_tiempo para el LATERAL de v_tiempos_vs_estandar'
EXPLAIN
SELECT utilizacion_meta_pct
FROM estandares.estandar_tiempo s
WHERE s.id_empresa = (SELECT id FROM gobierno.empresa LIMIT 1)
  AND s.id_tipo_equipo = (SELECT id FROM catalogos.tipo_equipo LIMIT 1)
  AND s.eliminado_en IS NULL
ORDER BY s.vigente_desde DESC
LIMIT 1;


-- 6E: consumo_explosivo — filtros por obra+tipo (informe por obra)
\echo '6E: EXPLAIN consumo_explosivo por obra + tipo_explosivo'
EXPLAIN
SELECT ce.id_tipo_explosivo, sum(ce.cantidad)
FROM produccion.consumo_explosivo ce
WHERE ce.id_obra = (SELECT id FROM catalogos.obra LIMIT 1)
GROUP BY ce.id_tipo_explosivo;


-- 6F: parte_acarreo — filtro por id_obra (patrón v_obras_prioritarias)
\echo '6F: EXPLAIN parte_acarreo por id_obra (subcorrelada en v_obras_prioritarias)'
EXPLAIN
SELECT pa.id, pa.fecha
FROM produccion.parte_acarreo pa
WHERE pa.id_obra = (SELECT id FROM catalogos.obra LIMIT 1)
  AND pa.eliminado_en IS NULL;


-- 6G: parte_barrenacion — filtro por id_obra (ídem, segunda subcorrelada)
\echo '6G: EXPLAIN parte_barrenacion por id_obra'
EXPLAIN
SELECT pb.id
FROM produccion.parte_barrenacion pb
WHERE pb.id_obra = (SELECT id FROM catalogos.obra LIMIT 1)
  AND pb.eliminado_en IS NULL;


\echo ''
\echo '=================================================================='
\echo 'BLOQUE 7: ÍNDICES SOBRE COLUMNAS DE AUDITORÍA (creado_por_usuario_id)'
\echo '=================================================================='

-- Las tablas de cabecera tienen creado_por y actualizado_por UUID sin índice.
-- Consultas de auditoría ("qué hizo el usuario X") harán seq scan a escala.
SELECT
  n.nspname AS esquema,
  c.relname AS tabla,
  EXISTS (
    SELECT 1 FROM pg_attribute a
    WHERE a.attrelid = c.oid AND a.attname = 'creado_por_usuario_id' AND a.attnum > 0
  ) AS tiene_creado_por,
  EXISTS (
    SELECT 1 FROM pg_indexes ix
    WHERE ix.schemaname = n.nspname AND ix.tablename = c.relname
      AND ix.indexdef ILIKE '%creado_por_usuario_id%'
  ) AS tiene_indice_creado_por
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r'
  AND n.nspname NOT IN ('pg_catalog','information_schema','pg_toast')
ORDER BY n.nspname, c.relname;


\echo ''
\echo '=================================================================='
\echo 'FIN AUDITORÍA D_rendimiento.sql'
\echo '=================================================================='
