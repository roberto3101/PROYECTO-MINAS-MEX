-- ============================================================
-- F_dominio.sql  —  Consistencia de dominio minero cross-capacidad
-- Agente: QA-Dominio   Fecha: 2026-06-09
-- SOLO LECTURA: únicamente SELECT / BEGIN...ROLLBACK si se necesita.
-- Resultado esperado: 0 filas en cada CHECK = PASA
-- ============================================================

\echo '========================================================'
\echo 'F_DOMINIO — Batería de consistencia de dominio minero'
\echo '========================================================'

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F01 — Tenant cruzado: acarreo_viaje.id_empresa vs parte_acarreo'
\echo '========================================================'
-- Un viaje cuyo id_empresa no coincide con el de su parte padre es un bug grave
SELECT
    'F01' AS check_id,
    av.id          AS viaje_id,
    av.id_empresa  AS viaje_empresa,
    pa.id_empresa  AS parte_empresa
FROM produccion.acarreo_viaje av
JOIN produccion.parte_acarreo pa ON pa.id = av.id_parte_acarreo
WHERE av.id_empresa <> pa.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F02 — Tenant cruzado: rezagado_ciclo.id_empresa vs parte_rezagado'
\echo '========================================================'
SELECT
    'F02' AS check_id,
    rc.id         AS ciclo_id,
    rc.id_empresa AS ciclo_empresa,
    pr.id_empresa AS parte_empresa
FROM produccion.rezagado_ciclo rc
JOIN produccion.parte_rezagado pr ON pr.id = rc.id_parte_rezagado
WHERE rc.id_empresa <> pr.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F03 — Tenant cruzado: parte_acarreo — obra y equipo deben pertenecer a la misma empresa'
\echo '========================================================'
SELECT
    'F03' AS check_id,
    pa.id         AS parte_id,
    pa.id_empresa AS parte_empresa,
    o.id_empresa  AS obra_empresa,
    e.id_empresa  AS equipo_empresa
FROM produccion.parte_acarreo pa
JOIN catalogos.obra    o ON o.id = pa.id_obra
JOIN catalogos.equipo  e ON e.id = pa.id_equipo
WHERE pa.id_empresa <> o.id_empresa
   OR pa.id_empresa <> e.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F04 — Tenant cruzado: parte_barrenacion — obra y equipo misma empresa'
\echo '========================================================'
SELECT
    'F04' AS check_id,
    pb.id         AS parte_id,
    pb.id_empresa AS parte_empresa,
    o.id_empresa  AS obra_empresa,
    e.id_empresa  AS equipo_empresa
FROM produccion.parte_barrenacion pb
JOIN catalogos.obra   o ON o.id = pb.id_obra
JOIN catalogos.equipo e ON e.id = pb.id_equipo
WHERE pb.id_empresa <> o.id_empresa
   OR pb.id_empresa <> e.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F05 — Tenant cruzado: parte_rezagado — obra y equipo misma empresa'
\echo '========================================================'
SELECT
    'F05' AS check_id,
    pr.id         AS parte_id,
    pr.id_empresa AS parte_empresa,
    o.id_empresa  AS obra_empresa,
    e.id_empresa  AS equipo_empresa
FROM produccion.parte_rezagado pr
JOIN catalogos.obra   o ON o.id = pr.id_obra
JOIN catalogos.equipo e ON e.id = pr.id_equipo
WHERE pr.id_empresa <> o.id_empresa
   OR pr.id_empresa <> e.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F06 — Tenant cruzado: obra.id_empresa debe coincidir con su mina'
\echo '========================================================'
SELECT
    'F06' AS check_id,
    o.id          AS obra_id,
    o.codigo      AS obra_codigo,
    o.id_empresa  AS obra_empresa,
    m.id_empresa  AS mina_empresa
FROM catalogos.obra o
JOIN catalogos.mina m ON m.id = o.id_mina
WHERE o.id_empresa <> m.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F07 — Tenant cruzado: equipo.id_empresa debe coincidir con su mina'
\echo '========================================================'
SELECT
    'F07' AS check_id,
    e.id          AS equipo_id,
    e.codigo      AS equipo_codigo,
    e.id_empresa  AS equipo_empresa,
    m.id_empresa  AS mina_empresa
FROM catalogos.equipo e
JOIN catalogos.mina m ON m.id = e.id_mina
WHERE e.id_empresa <> m.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F08 — Tenant cruzado: empleado.id_empresa vs su mina'
\echo '========================================================'
SELECT
    'F08' AS check_id,
    emp.id            AS empleado_id,
    emp.numero_nomina,
    emp.id_empresa    AS emp_empresa,
    m.id_empresa      AS mina_empresa
FROM catalogos.empleado emp
JOIN catalogos.mina m ON m.id = emp.id_mina
WHERE emp.id_empresa <> m.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F09 — Tenant cruzado: rebaje.id_empresa vs su mina (y obra si existe)'
\echo '========================================================'
SELECT
    'F09' AS check_id,
    r.id          AS rebaje_id,
    r.codigo      AS rebaje_codigo,
    r.id_empresa  AS rebaje_empresa,
    m.id_empresa  AS mina_empresa,
    o.id_empresa  AS obra_empresa
FROM planeacion.rebaje r
JOIN catalogos.mina  m ON m.id = r.id_mina
LEFT JOIN catalogos.obra o ON o.id = r.id_obra
WHERE r.id_empresa <> m.id_empresa
   OR (r.id_obra IS NOT NULL AND r.id_empresa <> o.id_empresa);

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F10 — Tenant cruzado: bloque_programado vs su plan y rebaje'
\echo '========================================================'
SELECT
    'F10' AS check_id,
    bp.id          AS bloque_id,
    bp.codigo_bloque,
    bp.id_empresa  AS bloque_empresa,
    p.id_empresa   AS plan_empresa,
    r.id_empresa   AS rebaje_empresa
FROM planeacion.bloque_programado bp
JOIN planeacion.plan   p ON p.id = bp.id_plan
JOIN planeacion.rebaje r ON r.id = bp.id_rebaje
WHERE bp.id_empresa <> p.id_empresa
   OR bp.id_empresa <> r.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F11 — Tenant cruzado: meta_periodo vs su bloque_programado'
\echo '========================================================'
SELECT
    'F11' AS check_id,
    mp.id          AS meta_id,
    mp.id_empresa  AS meta_empresa,
    bp.id_empresa  AS bloque_empresa
FROM planeacion.meta_periodo mp
JOIN planeacion.bloque_programado bp ON bp.id = mp.id_bloque_programado
WHERE mp.id_empresa <> bp.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F12 — Tenant cruzado: segmento vs su mina (y rebaje si existe)'
\echo '========================================================'
SELECT
    'F12' AS check_id,
    s.id          AS segmento_id,
    s.codigo,
    s.id_empresa  AS seg_empresa,
    m.id_empresa  AS mina_empresa,
    r.id_empresa  AS rebaje_empresa
FROM reconciliacion.segmento s
JOIN catalogos.mina   m ON m.id = s.id_mina
LEFT JOIN planeacion.rebaje r ON r.id = s.id_rebaje
WHERE s.id_empresa <> m.id_empresa
   OR (s.id_rebaje IS NOT NULL AND s.id_empresa <> r.id_empresa);

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F13 — Tenant cruzado: medicion vs su segmento'
\echo '========================================================'
SELECT
    'F13' AS check_id,
    med.id          AS medicion_id,
    med.id_empresa  AS med_empresa,
    s.id_empresa    AS seg_empresa
FROM reconciliacion.medicion med
JOIN reconciliacion.segmento s ON s.id = med.id_segmento
WHERE med.id_empresa <> s.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F14 — Horómetros: parte_acarreo (confirma CHECK constraint con datos reales)'
\echo '========================================================'
-- El CHECK en la tabla debería garantizar esto; lo verificamos explícitamente
SELECT
    'F14' AS check_id,
    id,
    fecha,
    horometro_inicial,
    horometro_final,
    horometro_final - horometro_inicial AS diferencia
FROM produccion.parte_acarreo
WHERE horometro_final < horometro_inicial;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F15 — Horómetros: parte_rezagado'
\echo '========================================================'
SELECT
    'F15' AS check_id,
    id,
    fecha,
    horometro_inicial,
    horometro_final,
    horometro_final - horometro_inicial AS diferencia
FROM produccion.parte_rezagado
WHERE horometro_final < horometro_inicial;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F16 — Horómetros: parte_barrenacion — diesel coherente cuando ambos no son NULL'
\echo '========================================================'
SELECT
    'F16' AS check_id,
    'diesel' AS par,
    id,
    fecha,
    horometro_diesel_inicial,
    horometro_diesel_final
FROM produccion.parte_barrenacion
WHERE horometro_diesel_inicial IS NOT NULL
  AND horometro_diesel_final   IS NOT NULL
  AND horometro_diesel_final < horometro_diesel_inicial;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F17 — Horómetros: parte_barrenacion — electrico coherente'
\echo '========================================================'
SELECT
    'F17' AS check_id,
    'electrico' AS par,
    id,
    fecha,
    horometro_electrico_inicial,
    horometro_electrico_final
FROM produccion.parte_barrenacion
WHERE horometro_electrico_inicial IS NOT NULL
  AND horometro_electrico_final   IS NOT NULL
  AND horometro_electrico_final < horometro_electrico_inicial;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F18 — Horómetros: parte_barrenacion — percusion coherente'
\echo '========================================================'
SELECT
    'F18' AS check_id,
    'percusion' AS par,
    id,
    fecha,
    horometro_percusion_inicial,
    horometro_percusion_final
FROM produccion.parte_barrenacion
WHERE horometro_percusion_inicial IS NOT NULL
  AND horometro_percusion_final   IS NOT NULL
  AND horometro_percusion_final < horometro_percusion_inicial;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F19 — Mineral vs tepetate: acarreo_viaje apunta a tipo_mineral activo'
\echo '========================================================'
-- El id_tipo_mineral en el viaje debe existir y estar ACTIVO (no INACTIVO)
SELECT
    'F19' AS check_id,
    av.id           AS viaje_id,
    av.id_tipo_mineral,
    tm.codigo       AS tipo_codigo,
    tm.es_mineral,
    tm.estado
FROM produccion.acarreo_viaje av
JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
WHERE tm.estado <> 'ACTIVO';

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F20 — Mineral vs tepetate: rezagado_ciclo apunta a tipo_mineral activo'
\echo '========================================================'
SELECT
    'F20' AS check_id,
    rc.id           AS ciclo_id,
    rc.id_tipo_mineral,
    tm.codigo       AS tipo_codigo,
    tm.es_mineral,
    tm.estado
FROM produccion.rezagado_ciclo rc
JOIN catalogos.tipo_mineral tm ON tm.id = rc.id_tipo_mineral
WHERE tm.estado <> 'ACTIVO';

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F21 — Mineral vs tepetate: tipo_mineral cross-tenant (viaje usa TM de otra empresa)'
\echo '========================================================'
SELECT
    'F21' AS check_id,
    av.id           AS viaje_id,
    av.id_empresa   AS viaje_empresa,
    tm.id_empresa   AS tm_empresa,
    tm.codigo
FROM produccion.acarreo_viaje av
JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
WHERE av.id_empresa <> tm.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F22 — Mineral vs tepetate: rezagado usa TM de otra empresa'
\echo '========================================================'
SELECT
    'F22' AS check_id,
    rc.id           AS ciclo_id,
    rc.id_empresa   AS ciclo_empresa,
    tm.id_empresa   AS tm_empresa,
    tm.codigo
FROM produccion.rezagado_ciclo rc
JOIN catalogos.tipo_mineral tm ON tm.id = rc.id_tipo_mineral
WHERE rc.id_empresa <> tm.id_empresa;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F23 — Mineral vs tepetate: balance por parte_acarreo (detalle de distribución)'
\echo '========================================================'
-- Informativo: muestra cuántas toneladas son mineral y cuántas tepetate por parte
SELECT
    'F23-INFO' AS check_id,
    pa.id            AS parte_id,
    pa.fecha,
    SUM(CASE WHEN tm.es_mineral = TRUE  THEN COALESCE(av.toneladas, 0) ELSE 0 END) AS ton_mineral,
    SUM(CASE WHEN tm.es_mineral = FALSE THEN COALESCE(av.toneladas, 0) ELSE 0 END) AS ton_tepetate,
    SUM(COALESCE(av.toneladas, 0))                                                   AS ton_total,
    COUNT(*)                                                                         AS n_viajes
FROM produccion.parte_acarreo pa
JOIN produccion.acarreo_viaje  av ON av.id_parte_acarreo = pa.id
JOIN catalogos.tipo_mineral    tm ON tm.id = av.id_tipo_mineral
GROUP BY pa.id, pa.fecha
ORDER BY pa.fecha;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F24 — Leyes: recuperacion_pct entre 0 y 100 (confirma CHECK constraint con datos)'
\echo '========================================================'
SELECT
    'F24' AS check_id,
    id,
    id_lote_molienda,
    metal,
    recuperacion_pct
FROM beneficio.recuperacion
WHERE recuperacion_pct IS NOT NULL
  AND (recuperacion_pct < 0 OR recuperacion_pct > 100);

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F25 — Leyes: ley_au no negativa en bloque_programado'
\echo '========================================================'
SELECT
    'F25' AS check_id,
    id,
    codigo_bloque,
    ley_au, ley_ag, ley_pb, ley_zn
FROM planeacion.bloque_programado
WHERE (ley_au IS NOT NULL AND ley_au < 0)
   OR (ley_ag IS NOT NULL AND ley_ag < 0)
   OR (ley_pb IS NOT NULL AND ley_pb < 0)
   OR (ley_zn IS NOT NULL AND ley_zn < 0);

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F26 — Leyes: leyes no negativas en medicion (reconciliacion)'
\echo '========================================================'
SELECT
    'F26' AS check_id,
    id,
    id_segmento,
    fuente,
    fecha,
    ley_au, ley_ag, ley_pb, ley_zn
FROM reconciliacion.medicion
WHERE (ley_au IS NOT NULL AND ley_au < 0)
   OR (ley_ag IS NOT NULL AND ley_ag < 0)
   OR (ley_pb IS NOT NULL AND ley_pb < 0)
   OR (ley_zn IS NOT NULL AND ley_zn < 0);

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F27 — Leyes: leyes no negativas en ley_metalurgica'
\echo '========================================================'
SELECT
    'F27' AS check_id,
    id,
    id_lote_molienda,
    punto,
    ley_au, ley_ag, ley_pb, ley_zn
FROM beneficio.ley_metalurgica
WHERE (ley_au IS NOT NULL AND ley_au < 0)
   OR (ley_ag IS NOT NULL AND ley_ag < 0)
   OR (ley_pb IS NOT NULL AND ley_pb < 0)
   OR (ley_zn IS NOT NULL AND ley_zn < 0);

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F28 — Reconciliacion: para SEG-2400W-001, verificar las 4 fuentes'
\echo '========================================================'
-- Debe existir exactamente 1 fila por fuente para el segmento de referencia
SELECT
    'F28-INFO' AS check_id,
    s.codigo   AS segmento,
    med.fuente,
    med.toneladas,
    med.ley_au
FROM reconciliacion.segmento s
JOIN reconciliacion.medicion med ON med.id_segmento = s.id
WHERE s.codigo = 'SEG-2400W-001'
ORDER BY
    CASE med.fuente
        WHEN 'RESERVA'    THEN 1
        WHEN 'TUMBE'      THEN 2
        WHEN 'ESTIMACION' THEN 3
        WHEN 'PLANTA'     THEN 4
    END;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F29 — Reconciliacion: cadena RESERVA >= TUMBE para SEG-2400W-001'
\echo '========================================================'
-- Si existen las fuentes, RESERVA >= TUMBE en toneladas tiene sentido geológico
SELECT
    'F29' AS check_id,
    s.codigo,
    res.toneladas AS ton_reserva,
    tum.toneladas AS ton_tumbe,
    CASE
        WHEN res.toneladas IS NULL THEN 'SIN_RESERVA'
        WHEN tum.toneladas IS NULL THEN 'SIN_TUMBE'
        WHEN res.toneladas >= tum.toneladas THEN 'PASA'
        ELSE 'FALLA — TUMBE > RESERVA'
    END AS resultado
FROM reconciliacion.segmento s
LEFT JOIN reconciliacion.medicion res
    ON res.id_segmento = s.id AND res.fuente = 'RESERVA'
LEFT JOIN reconciliacion.medicion tum
    ON tum.id_segmento = s.id AND tum.fuente = 'TUMBE'
WHERE s.codigo = 'SEG-2400W-001';

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F30 — Reconciliacion: fuentes disponibles para todos los segmentos activos'
\echo '========================================================'
-- Segmentos activos que no tienen ni una fuente registrada (advertencia)
SELECT
    'F30-WARN' AS check_id,
    s.id,
    s.codigo,
    s.estado
FROM reconciliacion.segmento s
LEFT JOIN reconciliacion.medicion med ON med.id_segmento = s.id
WHERE s.estado = 'ACTIVO'
  AND med.id IS NULL;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F31 — Plan: horizontes válidos (solo valores del conjunto definido)'
\echo '========================================================'
-- El CHECK constraint ya lo garantiza, lo verificamos con datos
SELECT
    'F31' AS check_id,
    horizonte,
    COUNT(*) AS n
FROM planeacion.meta_periodo
WHERE horizonte NOT IN ('ANUAL','SEMESTRAL','TRIMESTRAL','MENSUAL','SEMANAL','DIARIO')
GROUP BY horizonte;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F32 — Plan: suma mensual <= anual para el mismo bloque (misma empresa, misma etiqueta-año)'
\echo '========================================================'
-- Para cada bloque + año, la suma de todas las metas MENSUAL de ese año
-- no debe superar la meta ANUAL del mismo año.
-- Si no hay meta ANUAL no se puede validar (se reporta como advertencia).
WITH mensual AS (
    SELECT
        id_empresa,
        id_bloque_programado,
        LEFT(periodo_etiqueta, 4) AS anio,  -- 'YYYY' de 'YYYY-MM'
        SUM(COALESCE(toneladas_plan, 0))     AS suma_mensual
    FROM planeacion.meta_periodo
    WHERE horizonte = 'MENSUAL'
      AND periodo_etiqueta ~ '^\d{4}-\d{2}$'
    GROUP BY id_empresa, id_bloque_programado, LEFT(periodo_etiqueta, 4)
),
anual AS (
    SELECT
        id_empresa,
        id_bloque_programado,
        periodo_etiqueta                     AS anio,
        toneladas_plan                       AS ton_anual
    FROM planeacion.meta_periodo
    WHERE horizonte = 'ANUAL'
      AND periodo_etiqueta ~ '^\d{4}$'
)
SELECT
    'F32' AS check_id,
    m.id_empresa,
    m.id_bloque_programado,
    m.anio,
    m.suma_mensual,
    a.ton_anual,
    CASE
        WHEN a.ton_anual IS NULL THEN 'SIN_META_ANUAL'
        WHEN m.suma_mensual <= a.ton_anual THEN 'PASA'
        ELSE 'FALLA — mensual > anual'
    END AS resultado
FROM mensual m
LEFT JOIN anual a
    ON a.id_empresa = m.id_empresa
   AND a.id_bloque_programado = m.id_bloque_programado
   AND a.anio = m.anio
WHERE a.ton_anual IS NULL
   OR m.suma_mensual > a.ton_anual;  -- sólo retorna filas con problema o sin meta anual

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F33 — Plan: periodo_etiqueta con formato correcto según horizonte'
\echo '========================================================'
SELECT
    'F33' AS check_id,
    horizonte,
    periodo_etiqueta,
    COUNT(*) AS n_filas
FROM planeacion.meta_periodo
WHERE
    (horizonte = 'MENSUAL'  AND periodo_etiqueta !~ '^\d{4}-\d{2}$')
 OR (horizonte = 'ANUAL'    AND periodo_etiqueta !~ '^\d{4}$')
 OR (horizonte = 'SEMANAL'  AND periodo_etiqueta !~ '^\d{4}-W\d{2}$')
 OR (horizonte = 'DIARIO'   AND periodo_etiqueta !~ '^\d{4}-\d{2}-\d{2}$')
GROUP BY horizonte, periodo_etiqueta;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F34 — Fechas: ninguna fecha de negocio en el futuro lejano (> hoy + 5 años)'
\echo '========================================================'
SELECT 'F34-parte_acarreo'        AS origen, id, fecha FROM produccion.parte_acarreo     WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
UNION ALL
SELECT 'F34-parte_barrenacion',         id, fecha FROM produccion.parte_barrenacion  WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
UNION ALL
SELECT 'F34-parte_rezagado',            id, fecha FROM produccion.parte_rezagado     WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
UNION ALL
SELECT 'F34-medicion',                  id, fecha FROM reconciliacion.medicion       WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
UNION ALL
SELECT 'F34-bloque_programado', id, fecha_inicio  FROM planeacion.bloque_programado  WHERE fecha_inicio > CURRENT_DATE + INTERVAL '5 years';

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F35 — Fechas: ninguna fecha de negocio anterior a 1990 (dato irreal)'
\echo '========================================================'
SELECT 'F35-parte_acarreo'    AS origen, id, fecha FROM produccion.parte_acarreo    WHERE fecha < '1990-01-01'
UNION ALL
SELECT 'F35-parte_barrenacion',     id, fecha FROM produccion.parte_barrenacion WHERE fecha < '1990-01-01'
UNION ALL
SELECT 'F35-parte_rezagado',        id, fecha FROM produccion.parte_rezagado    WHERE fecha < '1990-01-01'
UNION ALL
SELECT 'F35-medicion',              id, fecha FROM reconciliacion.medicion      WHERE fecha < '1990-01-01';

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F36 — Huérfanos lógicos: parte_acarreo sin ningún viaje'
\echo '========================================================'
SELECT
    'F36-WARN' AS check_id,
    pa.id,
    pa.fecha,
    pa.turno,
    pa.id_empresa
FROM produccion.parte_acarreo pa
LEFT JOIN produccion.acarreo_viaje av ON av.id_parte_acarreo = pa.id
WHERE av.id IS NULL
  AND pa.eliminado_en IS NULL;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F37 — Huérfanos lógicos: parte_barrenacion sin ningún avance'
\echo '========================================================'
SELECT
    'F37-WARN' AS check_id,
    pb.id,
    pb.fecha,
    pb.turno,
    pb.tipo_barrenacion,
    pb.id_empresa
FROM produccion.parte_barrenacion pb
LEFT JOIN produccion.barrenacion_avance ba ON ba.id_parte_barrenacion = pb.id
WHERE ba.id IS NULL
  AND pb.eliminado_en IS NULL;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F38 — Huérfanos lógicos: parte_rezagado sin ningún ciclo'
\echo '========================================================'
SELECT
    'F38-WARN' AS check_id,
    pr.id,
    pr.fecha,
    pr.turno,
    pr.id_empresa
FROM produccion.parte_rezagado pr
LEFT JOIN produccion.rezagado_ciclo rc ON rc.id_parte_rezagado = pr.id
WHERE rc.id IS NULL
  AND pr.eliminado_en IS NULL;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'CHECK F39 — Huérfanos lógicos: plan sin ningún bloque_programado'
\echo '========================================================'
SELECT
    'F39-WARN' AS check_id,
    p.id,
    p.nombre,
    p.anio,
    p.id_empresa
FROM planeacion.plan p
LEFT JOIN planeacion.bloque_programado bp ON bp.id_plan = p.id
WHERE bp.id IS NULL
  AND p.eliminado_en IS NULL;

-- ============================================================
\echo ''
\echo '========================================================'
\echo 'RESUMEN — Conteo de filas por check (0 en todos los F0x..F3x = perfecto)'
\echo '========================================================'
-- Ejecutamos conteos individuales en una CTE para el resumen global
WITH resultados(check_id, n) AS (
    -- F01
    SELECT 'F01', COUNT(*) FROM produccion.acarreo_viaje av JOIN produccion.parte_acarreo pa ON pa.id = av.id_parte_acarreo WHERE av.id_empresa <> pa.id_empresa
    UNION ALL
    -- F02
    SELECT 'F02', COUNT(*) FROM produccion.rezagado_ciclo rc JOIN produccion.parte_rezagado pr ON pr.id = rc.id_parte_rezagado WHERE rc.id_empresa <> pr.id_empresa
    UNION ALL
    -- F03
    SELECT 'F03', COUNT(*) FROM produccion.parte_acarreo pa JOIN catalogos.obra o ON o.id = pa.id_obra JOIN catalogos.equipo e ON e.id = pa.id_equipo WHERE pa.id_empresa <> o.id_empresa OR pa.id_empresa <> e.id_empresa
    UNION ALL
    -- F04
    SELECT 'F04', COUNT(*) FROM produccion.parte_barrenacion pb JOIN catalogos.obra o ON o.id = pb.id_obra JOIN catalogos.equipo e ON e.id = pb.id_equipo WHERE pb.id_empresa <> o.id_empresa OR pb.id_empresa <> e.id_empresa
    UNION ALL
    -- F05
    SELECT 'F05', COUNT(*) FROM produccion.parte_rezagado pr JOIN catalogos.obra o ON o.id = pr.id_obra JOIN catalogos.equipo e ON e.id = pr.id_equipo WHERE pr.id_empresa <> o.id_empresa OR pr.id_empresa <> e.id_empresa
    UNION ALL
    -- F06
    SELECT 'F06', COUNT(*) FROM catalogos.obra o JOIN catalogos.mina m ON m.id = o.id_mina WHERE o.id_empresa <> m.id_empresa
    UNION ALL
    -- F07
    SELECT 'F07', COUNT(*) FROM catalogos.equipo e JOIN catalogos.mina m ON m.id = e.id_mina WHERE e.id_empresa <> m.id_empresa
    UNION ALL
    -- F08
    SELECT 'F08', COUNT(*) FROM catalogos.empleado emp JOIN catalogos.mina m ON m.id = emp.id_mina WHERE emp.id_empresa <> m.id_empresa
    UNION ALL
    -- F09
    SELECT 'F09', COUNT(*) FROM planeacion.rebaje r JOIN catalogos.mina m ON m.id = r.id_mina LEFT JOIN catalogos.obra o ON o.id = r.id_obra WHERE r.id_empresa <> m.id_empresa OR (r.id_obra IS NOT NULL AND r.id_empresa <> o.id_empresa)
    UNION ALL
    -- F10
    SELECT 'F10', COUNT(*) FROM planeacion.bloque_programado bp JOIN planeacion.plan p ON p.id = bp.id_plan JOIN planeacion.rebaje r ON r.id = bp.id_rebaje WHERE bp.id_empresa <> p.id_empresa OR bp.id_empresa <> r.id_empresa
    UNION ALL
    -- F11
    SELECT 'F11', COUNT(*) FROM planeacion.meta_periodo mp JOIN planeacion.bloque_programado bp ON bp.id = mp.id_bloque_programado WHERE mp.id_empresa <> bp.id_empresa
    UNION ALL
    -- F12
    SELECT 'F12', COUNT(*) FROM reconciliacion.segmento s JOIN catalogos.mina m ON m.id = s.id_mina LEFT JOIN planeacion.rebaje r ON r.id = s.id_rebaje WHERE s.id_empresa <> m.id_empresa OR (s.id_rebaje IS NOT NULL AND s.id_empresa <> r.id_empresa)
    UNION ALL
    -- F13
    SELECT 'F13', COUNT(*) FROM reconciliacion.medicion med JOIN reconciliacion.segmento s ON s.id = med.id_segmento WHERE med.id_empresa <> s.id_empresa
    UNION ALL
    -- F14
    SELECT 'F14', COUNT(*) FROM produccion.parte_acarreo WHERE horometro_final < horometro_inicial
    UNION ALL
    -- F15
    SELECT 'F15', COUNT(*) FROM produccion.parte_rezagado WHERE horometro_final < horometro_inicial
    UNION ALL
    -- F16
    SELECT 'F16', COUNT(*) FROM produccion.parte_barrenacion WHERE horometro_diesel_inicial IS NOT NULL AND horometro_diesel_final IS NOT NULL AND horometro_diesel_final < horometro_diesel_inicial
    UNION ALL
    -- F17
    SELECT 'F17', COUNT(*) FROM produccion.parte_barrenacion WHERE horometro_electrico_inicial IS NOT NULL AND horometro_electrico_final IS NOT NULL AND horometro_electrico_final < horometro_electrico_inicial
    UNION ALL
    -- F18
    SELECT 'F18', COUNT(*) FROM produccion.parte_barrenacion WHERE horometro_percusion_inicial IS NOT NULL AND horometro_percusion_final IS NOT NULL AND horometro_percusion_final < horometro_percusion_inicial
    UNION ALL
    -- F19
    SELECT 'F19', COUNT(*) FROM produccion.acarreo_viaje av JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral WHERE tm.estado <> 'ACTIVO'
    UNION ALL
    -- F20
    SELECT 'F20', COUNT(*) FROM produccion.rezagado_ciclo rc JOIN catalogos.tipo_mineral tm ON tm.id = rc.id_tipo_mineral WHERE tm.estado <> 'ACTIVO'
    UNION ALL
    -- F21
    SELECT 'F21', COUNT(*) FROM produccion.acarreo_viaje av JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral WHERE av.id_empresa <> tm.id_empresa
    UNION ALL
    -- F22
    SELECT 'F22', COUNT(*) FROM produccion.rezagado_ciclo rc JOIN catalogos.tipo_mineral tm ON tm.id = rc.id_tipo_mineral WHERE rc.id_empresa <> tm.id_empresa
    UNION ALL
    -- F24
    SELECT 'F24', COUNT(*) FROM beneficio.recuperacion WHERE recuperacion_pct IS NOT NULL AND (recuperacion_pct < 0 OR recuperacion_pct > 100)
    UNION ALL
    -- F25
    SELECT 'F25', COUNT(*) FROM planeacion.bloque_programado WHERE (ley_au IS NOT NULL AND ley_au < 0) OR (ley_ag IS NOT NULL AND ley_ag < 0) OR (ley_pb IS NOT NULL AND ley_pb < 0) OR (ley_zn IS NOT NULL AND ley_zn < 0)
    UNION ALL
    -- F26
    SELECT 'F26', COUNT(*) FROM reconciliacion.medicion WHERE (ley_au IS NOT NULL AND ley_au < 0) OR (ley_ag IS NOT NULL AND ley_ag < 0) OR (ley_pb IS NOT NULL AND ley_pb < 0) OR (ley_zn IS NOT NULL AND ley_zn < 0)
    UNION ALL
    -- F27
    SELECT 'F27', COUNT(*) FROM beneficio.ley_metalurgica WHERE (ley_au IS NOT NULL AND ley_au < 0) OR (ley_ag IS NOT NULL AND ley_ag < 0) OR (ley_pb IS NOT NULL AND ley_pb < 0) OR (ley_zn IS NOT NULL AND ley_zn < 0)
    UNION ALL
    -- F31
    SELECT 'F31', COUNT(*) FROM planeacion.meta_periodo WHERE horizonte NOT IN ('ANUAL','SEMESTRAL','TRIMESTRAL','MENSUAL','SEMANAL','DIARIO')
    UNION ALL
    -- F33
    SELECT 'F33', COUNT(*) FROM planeacion.meta_periodo WHERE (horizonte = 'MENSUAL' AND periodo_etiqueta !~ '^\d{4}-\d{2}$') OR (horizonte = 'ANUAL' AND periodo_etiqueta !~ '^\d{4}$') OR (horizonte = 'SEMANAL' AND periodo_etiqueta !~ '^\d{4}-W\d{2}$') OR (horizonte = 'DIARIO' AND periodo_etiqueta !~ '^\d{4}-\d{2}-\d{2}$')
    UNION ALL
    -- F34
    SELECT 'F34', COUNT(*) FROM (
        SELECT id FROM produccion.parte_acarreo    WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
        UNION ALL SELECT id FROM produccion.parte_barrenacion WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
        UNION ALL SELECT id FROM produccion.parte_rezagado    WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
        UNION ALL SELECT id FROM reconciliacion.medicion      WHERE fecha > CURRENT_DATE + INTERVAL '5 years'
    ) t
    UNION ALL
    -- F35
    SELECT 'F35', COUNT(*) FROM (
        SELECT id FROM produccion.parte_acarreo    WHERE fecha < '1990-01-01'
        UNION ALL SELECT id FROM produccion.parte_barrenacion WHERE fecha < '1990-01-01'
        UNION ALL SELECT id FROM produccion.parte_rezagado    WHERE fecha < '1990-01-01'
        UNION ALL SELECT id FROM reconciliacion.medicion      WHERE fecha < '1990-01-01'
    ) t
)
SELECT
    check_id,
    n                                                    AS filas_inconsistentes,
    CASE WHEN n = 0 THEN 'PASA' ELSE 'FALLA' END        AS resultado
FROM resultados
ORDER BY check_id;
