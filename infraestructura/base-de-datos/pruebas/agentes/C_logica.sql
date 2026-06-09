-- ============================================================
-- C_logica.sql  · Agente C: Lógica de negocio y vistas
-- Schema: reportes  — SOLO LECTURA (SELECT + REFRESH MV)
-- Ejecutar: psql -U postgres -h 127.0.0.1 -p 5433 -d mina -f C_logica.sql
-- ============================================================

\set ON_ERROR_STOP off
\timing on

-- Constantes de filtro (IDs del seed)
\set m_cien  'aaaa0000-0000-0000-0000-000000000001'
\set emp     '11111111-1111-1111-1111-111111111111'

\echo ''
\echo '============================================================'
\echo 'PASO 0 — Refrescar vista materializada (idempotente/seguro)'
\echo '============================================================'
REFRESH MATERIALIZED VIEW reportes.mv_plan_vs_real_mensual;

\echo ''
\echo '============================================================'
\echo 'BLOQUE 1 — v_balance_cargas (La Cienega, 2026-06)'
\echo '  Esperado: ton_mineral=760, ton_tepetate=200, ton_total=960'
\echo '============================================================'

-- Test 1-A: verificar valores puntuales
SELECT
    periodo_mes,
    ton_mineral,
    ton_tepetate,
    ton_total,
    CASE WHEN ton_mineral  = 760  THEN 'PASA' ELSE 'FALLA ton_mineral='  || ton_mineral  END AS chk_mineral,
    CASE WHEN ton_tepetate = 200  THEN 'PASA' ELSE 'FALLA ton_tepetate=' || ton_tepetate END AS chk_tepetate,
    CASE WHEN ton_total    = 960  THEN 'PASA' ELSE 'FALLA ton_total='    || ton_total    END AS chk_total
FROM reportes.v_balance_cargas
WHERE id_mina = :'m_cien'::uuid
  AND periodo_mes = '2026-06';

-- Test 1-B: coherencia interna ton_mineral + ton_tepetate = ton_total
SELECT
    periodo_mes,
    ton_mineral + ton_tepetate AS suma_partes,
    ton_total,
    CASE WHEN (ton_mineral + ton_tepetate) = ton_total
         THEN 'PASA coherencia interna'
         ELSE 'FALLA: suma_partes=' || (ton_mineral+ton_tepetate) || ' <> ton_total=' || ton_total
    END AS coherencia
FROM reportes.v_balance_cargas
WHERE id_mina = :'m_cien'::uuid;

-- Test 1-C: borde — minas sin datos deben devolver 0 (no NULL)
SELECT
    mina,
    ton_mineral,
    ton_tepetate,
    ton_total,
    CASE WHEN ton_mineral  IS NULL THEN 'FALLA NULL ton_mineral'
         WHEN ton_tepetate IS NULL THEN 'FALLA NULL ton_tepetate'
         WHEN ton_total    IS NULL THEN 'FALLA NULL ton_total'
         ELSE 'PASA (no NULLs)'
    END AS chk_nulls
FROM reportes.v_balance_cargas;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 2 — mv_plan_vs_real_mensual (La Cienega, 2026-06)'
\echo '  Esperado: ton_plan=1000, ton_real=760, dif=-240, apego=76.00'
\echo '============================================================'

-- Test 2-A: valores puntuales
SELECT
    periodo,
    ton_plan,
    ton_real,
    diferencia,
    apego_pct,
    CASE WHEN ton_plan   = 1000   THEN 'PASA' ELSE 'FALLA ton_plan='   || ton_plan   END AS chk_plan,
    CASE WHEN ton_real   = 760    THEN 'PASA' ELSE 'FALLA ton_real='   || ton_real   END AS chk_real,
    CASE WHEN diferencia = -240   THEN 'PASA' ELSE 'FALLA dif='        || diferencia END AS chk_dif,
    CASE WHEN apego_pct  = 76.00  THEN 'PASA' ELSE 'FALLA apego_pct='  || apego_pct  END AS chk_apego
FROM reportes.mv_plan_vs_real_mensual
WHERE id_mina = :'m_cien'::uuid
  AND periodo = '2026-06';

-- Test 2-B: protección división por cero (plan=0 => apego_pct NULL, no explosion)
-- Verifica que no hay filas con ton_plan=0 y apego_pct NOT NULL
SELECT
    periodo,
    ton_plan,
    apego_pct,
    CASE WHEN ton_plan = 0 AND apego_pct IS NOT NULL
         THEN 'FALLA: division por cero no protegida (apego_pct=' || apego_pct || ')'
         ELSE 'PASA (NULLIF protege div/0)'
    END AS chk_div0
FROM reportes.mv_plan_vs_real_mensual;

-- Test 2-C: diferencia = ton_real - ton_plan (coherencia fórmula)
SELECT
    periodo,
    ton_real - ton_plan          AS dif_calculada,
    diferencia                   AS dif_guardada,
    CASE WHEN (ton_real - ton_plan) = diferencia
         THEN 'PASA coherencia diferencia'
         ELSE 'FALLA: calculada=' || (ton_real-ton_plan) || ' <> guardada=' || diferencia
    END AS chk_formula
FROM reportes.mv_plan_vs_real_mensual
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 3 — v_avance_barrenacion (La Cienega, 2026-06-01)'
\echo '  Esperado: metros=20, barrenos=10'
\echo '============================================================'

SELECT
    fecha,
    obra,
    metros,
    barrenos,
    CASE WHEN metros   = 20 THEN 'PASA' ELSE 'FALLA metros='   || metros   END AS chk_metros,
    CASE WHEN barrenos = 10 THEN 'PASA' ELSE 'FALLA barrenos=' || barrenos END AS chk_barrenos
FROM reportes.v_avance_barrenacion
WHERE id_mina = :'m_cien'::uuid
  AND fecha   = '2026-06-01';

-- Test 3-B: sin NULLs en columnas calculadas
SELECT
    fecha, obra, metros, barrenos,
    CASE WHEN metros   IS NULL THEN 'FALLA NULL metros'
         WHEN barrenos IS NULL THEN 'FALLA NULL barrenos'
         ELSE 'PASA (no NULLs)'
    END AS chk_nulls
FROM reportes.v_avance_barrenacion;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 4 — v_reconciliacion (SEG-2400W-001)'
\echo '  Esperado: dif_tumbe_reserva=-50, dif_planta_tumbe=-50'
\echo '  Seed: RESERVA=1000, TUMBE=950, ESTIMACION=940, PLANTA=900'
\echo '  dif_tumbe_reserva = 950-1000 = -50  ✓'
\echo '  dif_planta_tumbe  = 900-950  = -50  ✓'
\echo '============================================================'

SELECT
    segmento,
    ton_reserva,
    ton_tumbe,
    ton_estimacion,
    ton_planta,
    dif_tumbe_reserva,
    dif_planta_tumbe,
    CASE WHEN dif_tumbe_reserva = -50 THEN 'PASA' ELSE 'FALLA dif_tumbe_reserva=' || dif_tumbe_reserva END AS chk_t_r,
    CASE WHEN dif_planta_tumbe  = -50 THEN 'PASA' ELSE 'FALLA dif_planta_tumbe='  || dif_planta_tumbe  END AS chk_p_t
FROM reportes.v_reconciliacion
WHERE segmento = 'SEG-2400W-001';

-- Test 4-B: segmentos sin mediciones deben devolver NULL (FILTER no fuerza 0, es correcto)
--           pero verificamos que la vista no rompe con segmentos sin datos
SELECT count(*) AS total_segmentos_visible
FROM reportes.v_reconciliacion;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 5 — v_disponibilidad_equipo (EQ-001, 2026-06-01)'
\echo '  Seed: pa1 horometro 5420.5→5428 = 7.5h; demora EQ-001 = 60min = 1.0h'
\echo '  Esperado: horas_operativas=7.5, horas_demora=1.0'
\echo '  utilizacion = 7.5/(7.5+1.0)*100 = 7.5/8.5*100 = 88.235... ≈ 88.24'
\echo '============================================================'

SELECT
    equipo,
    fecha,
    horas_operativas,
    horas_demora,
    utilizacion_pct,
    CASE WHEN horas_operativas = 7.5   THEN 'PASA' ELSE 'FALLA horas_op='  || horas_operativas END AS chk_hop,
    CASE WHEN horas_demora     = 1.0   THEN 'PASA' ELSE 'FALLA horas_dem=' || horas_demora     END AS chk_hdem,
    CASE WHEN utilizacion_pct  = 88.24 THEN 'PASA' ELSE 'FALLA util_pct='  || utilizacion_pct  END AS chk_util
FROM reportes.v_disponibilidad_equipo
WHERE equipo = 'EQ-001'
  AND fecha  = '2026-06-01';

-- Test 5-B: protección NULLIF (divisor nunca 0 si hay horas)
SELECT
    equipo, fecha, horas_operativas, utilizacion_pct,
    CASE WHEN horas_operativas > 0 AND utilizacion_pct IS NULL
         THEN 'FALLA: horas>0 pero utilizacion_pct NULL (NULLIF mal puesto)'
         ELSE 'PASA div/0 protegida'
    END AS chk_nullif
FROM reportes.v_disponibilidad_equipo;

-- Test 5-C: equipo sin horas (hipotético) => utilizacion_pct debe ser NULL, no crash
--           Como no hay parte_acarreo para EQ-002 en otra fecha, este test es informativo
SELECT count(*) AS filas_disponibilidad FROM reportes.v_disponibilidad_equipo;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 6 — v_tiempos_vs_estandar (EQ-001)'
\echo '  Esperado: utilizacion_meta_pct=80 (estándar CAMION)'
\echo '============================================================'

SELECT
    equipo,
    tipo_equipo,
    fecha,
    horas_operativas,
    utilizacion_pct,
    utilizacion_meta_pct,
    desvio_utilizacion_pct,
    CASE WHEN utilizacion_meta_pct = 80 THEN 'PASA' ELSE 'FALLA meta_pct=' || utilizacion_meta_pct END AS chk_meta,
    CASE WHEN desvio_utilizacion_pct = round(utilizacion_pct - 80, 2)
         THEN 'PASA desvio correcto'
         ELSE 'FALLA desvio=' || desvio_utilizacion_pct || ' esperado=' || round(utilizacion_pct-80,2)
    END AS chk_desvio
FROM reportes.v_tiempos_vs_estandar
WHERE equipo = 'EQ-001';

-- Test 6-B: tipo_equipo sin estándar => utilizacion_meta_pct NULL (LEFT LATERAL correcto)
SELECT
    equipo, tipo_equipo, utilizacion_meta_pct,
    CASE WHEN utilizacion_meta_pct IS NULL
         THEN 'INFO: sin estándar (LEFT LATERAL ok, no crash)'
         ELSE 'PASA: tiene estándar=' || utilizacion_meta_pct
    END AS chk_sin_estandar
FROM reportes.v_tiempos_vs_estandar;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 7 — v_produccion_vs_molienda (La Cienega, 2026-06)'
\echo '  Esperado: ton_acarreada=760, ton_molida=720, diferencia=-40'
\echo '  Seed: pa1=400+360=760 mineral; lote1=720 toneladas_molidas'
\echo '  diferencia = 720-760 = -40'
\echo '============================================================'

SELECT
    periodo,
    ton_acarreada,
    ton_molida,
    diferencia,
    CASE WHEN ton_acarreada = 760  THEN 'PASA' ELSE 'FALLA ton_acarreada=' || ton_acarreada END AS chk_acar,
    CASE WHEN ton_molida    = 720  THEN 'PASA' ELSE 'FALLA ton_molida='    || ton_molida    END AS chk_mol,
    CASE WHEN diferencia    = -40  THEN 'PASA' ELSE 'FALLA diferencia='    || diferencia    END AS chk_dif
FROM reportes.v_produccion_vs_molienda
WHERE id_mina  = :'m_cien'::uuid
  AND periodo  = '2026-06';

-- Test 7-B: coherencia diferencia = ton_molida - ton_acarreada
SELECT
    periodo,
    ton_molida - ton_acarreada AS calculada,
    diferencia                 AS guardada,
    CASE WHEN (ton_molida - ton_acarreada) = diferencia
         THEN 'PASA coherencia'
         ELSE 'FALLA: calc=' || (ton_molida-ton_acarreada) || ' <> guard=' || diferencia
    END AS chk_formula
FROM reportes.v_produccion_vs_molienda
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 8 — v_balance_metalurgico (LOTE-2026-06-01)'
\echo '  Esperado: rec_au_pct=92.5, au_cabeza=0.78'
\echo '============================================================'

SELECT
    codigo_lote,
    fecha,
    toneladas_molidas,
    au_cabeza,
    au_concentrado,
    au_cola,
    rec_au_pct,
    rec_ag_pct,
    CASE WHEN rec_au_pct = 92.5 THEN 'PASA' ELSE 'FALLA rec_au_pct=' || rec_au_pct END AS chk_rec_au,
    CASE WHEN au_cabeza  = 0.78 THEN 'PASA' ELSE 'FALLA au_cabeza='  || au_cabeza  END AS chk_au_cab
FROM reportes.v_balance_metalurgico
WHERE codigo_lote = 'LOTE-2026-06-01';

-- Test 8-B: sin NULLs inesperados en columnas de leyes cuando hay datos
SELECT
    codigo_lote,
    au_cabeza, au_concentrado, au_cola, rec_au_pct, rec_ag_pct,
    CASE WHEN au_cabeza     IS NULL THEN 'FALLA NULL au_cabeza'
         WHEN au_concentrado IS NULL THEN 'FALLA NULL au_concentrado'
         WHEN rec_au_pct    IS NULL THEN 'FALLA NULL rec_au_pct'
         ELSE 'PASA (no NULLs críticos)'
    END AS chk_nulls
FROM reportes.v_balance_metalurgico
WHERE codigo_lote = 'LOTE-2026-06-01';


\echo ''
\echo '============================================================'
\echo 'BLOQUE 9 — v_obras_prioritarias (REB-2400W)'
\echo '  Esperado: ton_mineral=760, metros_barrenados=20'
\echo '============================================================'

SELECT
    obra,
    nombre,
    ton_mineral,
    metros_barrenados,
    CASE WHEN ton_mineral      = 760 THEN 'PASA' ELSE 'FALLA ton_mineral='     || ton_mineral      END AS chk_ton,
    CASE WHEN metros_barrenados = 20 THEN 'PASA' ELSE 'FALLA metros_barren='   || metros_barrenados END AS chk_metros
FROM reportes.v_obras_prioritarias
WHERE obra = 'REB-2400W';

-- Test 9-B: obras no prioritarias NO aparecen en la vista
SELECT
    count(*) AS total_prioritarias,
    CASE WHEN count(*) = 1 THEN 'PASA (solo 1 obra prioritaria en seed)'
         ELSE 'INFO: ' || count(*) || ' obras prioritarias'
    END AS chk_filtro
FROM reportes.v_obras_prioritarias;

-- Test 9-C: campos coalesce nunca NULL
SELECT
    obra, ton_mineral, metros_barrenados,
    CASE WHEN ton_mineral      IS NULL THEN 'FALLA NULL ton_mineral'
         WHEN metros_barrenados IS NULL THEN 'FALLA NULL metros_barrenados'
         ELSE 'PASA (COALESCE ok)'
    END AS chk_nulls
FROM reportes.v_obras_prioritarias;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 10 — v_inversion_obra (REB-2400W, 2026-06)'
\echo '  Esperado: inversion=120000'
\echo '============================================================'

SELECT
    obra,
    periodo,
    inversion,
    moneda,
    CASE WHEN inversion = 120000 THEN 'PASA' ELSE 'FALLA inversion=' || inversion END AS chk_inv
FROM reportes.v_inversion_obra
WHERE obra   = 'REB-2400W'
  AND periodo = '2026-06';

-- Test 10-B: también existe RPA-2140 con 80000
SELECT
    obra,
    periodo,
    inversion,
    CASE WHEN obra='RPA-2140' AND inversion=80000 THEN 'PASA RPA=80000'
         WHEN obra='REB-2400W' AND inversion=120000 THEN 'PASA REB=120000'
         ELSE 'FALLA obra=' || obra || ' inv=' || inversion
    END AS chk
FROM reportes.v_inversion_obra
WHERE periodo = '2026-06'
ORDER BY obra;

-- Test 10-C: no NULLs en inversion (COALESCE)
SELECT
    obra, inversion,
    CASE WHEN inversion IS NULL THEN 'FALLA NULL inversion' ELSE 'PASA COALESCE ok' END AS chk_null
FROM reportes.v_inversion_obra;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 11 — v_presupuesto_vs_real (La Cienega, 2026-06)'
\echo '  Seed: presupuesto=50000, estimacion_conceptos=30000+15000=45000'
\echo '  variacion = real - presupuesto = 45000-50000 = -5000'
\echo '  Esperado: variacion=-5000'
\echo '============================================================'

SELECT
    periodo,
    presupuesto,
    real_importe,
    variacion,
    CASE WHEN presupuesto  = 50000 THEN 'PASA' ELSE 'FALLA presupuesto='  || presupuesto  END AS chk_pres,
    CASE WHEN real_importe = 45000 THEN 'PASA' ELSE 'FALLA real_importe=' || real_importe END AS chk_real,
    CASE WHEN variacion    = -5000 THEN 'PASA' ELSE 'FALLA variacion='    || variacion    END AS chk_var
FROM reportes.v_presupuesto_vs_real
WHERE id_mina  = :'m_cien'::uuid
  AND periodo  = '2026-06';

-- Test 11-B: coherencia variacion = real - presupuesto
SELECT
    periodo,
    real_importe - presupuesto AS calculada,
    variacion                  AS guardada,
    CASE WHEN (real_importe - presupuesto) = variacion
         THEN 'PASA coherencia variacion'
         ELSE 'FALLA: calc=' || (real_importe-presupuesto) || ' <> guard=' || variacion
    END AS chk_formula
FROM reportes.v_presupuesto_vs_real
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 12 — v_costo_unitario (filas coherentes, sin NULLs)'
\echo '  Seed: 2 filas para La Cienega 2026-06 (Acarreo=2.35, Desarrollo=310.50)'
\echo '============================================================'

SELECT
    periodo,
    actividad,
    area,
    unidad,
    valor,
    moneda,
    CASE WHEN valor  IS NULL  THEN 'FALLA NULL valor'
         WHEN unidad IS NULL  THEN 'FALLA NULL unidad'
         WHEN actividad IS NULL THEN 'FALLA NULL actividad'
         ELSE 'PASA'
    END AS chk_nulls
FROM reportes.v_costo_unitario
WHERE id_mina  = :'m_cien'::uuid
  AND periodo  = '2026-06'
ORDER BY actividad;

-- Test 12-B: valores específicos
SELECT
    count(*) AS total_filas,
    CASE WHEN count(*) = 2 THEN 'PASA (2 costos unitarios en seed)'
         ELSE 'FALLA: esperado 2, obtenido ' || count(*)
    END AS chk_count,
    max(CASE WHEN actividad='Acarreo'    THEN valor END) AS val_acarreo,
    max(CASE WHEN actividad='Desarrollo' THEN valor END) AS val_desarrollo
FROM reportes.v_costo_unitario
WHERE id_mina = :'m_cien'::uuid AND periodo = '2026-06';


\echo ''
\echo '============================================================'
\echo 'BLOQUE 13 — v_costo_estimacion (filas coherentes)'
\echo '  Seed: EST-2026-06-001 DESARROLLO 45000 (30000+15000)'
\echo '============================================================'

SELECT
    periodo,
    tipo,
    importe,
    CASE WHEN importe IS NULL THEN 'FALLA NULL importe'
         WHEN importe = 45000 THEN 'PASA importe=45000'
         ELSE 'FALLA importe=' || importe
    END AS chk
FROM reportes.v_costo_estimacion
WHERE id_mina  = :'m_cien'::uuid
  AND periodo  = '2026-06';

-- Test 13-B: COALESCE protege importe cuando no hay conceptos
SELECT
    periodo, tipo, importe,
    CASE WHEN importe IS NULL THEN 'FALLA NULL (COALESCE roto)' ELSE 'PASA COALESCE ok' END AS chk_null
FROM reportes.v_costo_estimacion;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 14 — v_consumo_explosivo (filas coherentes, no vacía)'
\echo '  Seed: ANFO 250 kg + FULMINANTE 10 pza, destino TUMBE, obra REB-2400W'
\echo '============================================================'

SELECT
    obra,
    destino,
    familia,
    periodo,
    cantidad,
    CASE WHEN cantidad IS NULL THEN 'FALLA NULL cantidad'
         WHEN cantidad <= 0    THEN 'FALLA cantidad<=0'
         ELSE 'PASA'
    END AS chk_cantidad
FROM reportes.v_consumo_explosivo
WHERE id_mina  = :'m_cien'::uuid
ORDER BY familia;

SELECT
    count(*) AS filas_explosivo,
    CASE WHEN count(*) = 2 THEN 'PASA (2 tipos en seed)'
         ELSE 'FALLA: esperado 2, obtenido ' || count(*)
    END AS chk_count
FROM reportes.v_consumo_explosivo
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 15 — v_plan_periodo (filas coherentes)'
\echo '  Seed: bp_819 MENSUAL 2026-06 => 1000 ton, ANUAL 2026 => 12000 ton'
\echo '============================================================'

SELECT
    horizonte,
    periodo_etiqueta,
    rebaje,
    toneladas_plan,
    CASE WHEN toneladas_plan IS NULL THEN 'FALLA NULL toneladas_plan'
         WHEN horizonte='MENSUAL' AND periodo_etiqueta='2026-06' AND toneladas_plan=1000  THEN 'PASA MENSUAL=1000'
         WHEN horizonte='ANUAL'   AND periodo_etiqueta='2026'    AND toneladas_plan=12000 THEN 'PASA ANUAL=12000'
         ELSE 'INFO: ' || horizonte || '/' || periodo_etiqueta || '=' || toneladas_plan
    END AS chk
FROM reportes.v_plan_periodo
WHERE id_mina = :'m_cien'::uuid
ORDER BY horizonte, periodo_etiqueta;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 16 — v_produccion_acarreo (filas coherentes, no vacía)'
\echo '  Seed: pa1 MINERAL=400 TEPETATE=120; pa2 MINERAL=360 TEPETATE=80'
\echo '============================================================'

SELECT
    fecha, turno, tipo_material, es_mineral, viajes, toneladas,
    CASE WHEN toneladas IS NULL THEN 'FALLA NULL toneladas'
         WHEN viajes    IS NULL THEN 'FALLA NULL viajes'
         ELSE 'PASA'
    END AS chk_nulls
FROM reportes.v_produccion_acarreo
WHERE id_mina = :'m_cien'::uuid
ORDER BY fecha, tipo_material;

-- Total mineral de acarreo = 400+360=760
SELECT
    sum(toneladas) FILTER (WHERE es_mineral)      AS total_mineral,
    sum(toneladas) FILTER (WHERE NOT es_mineral)  AS total_no_mineral,
    CASE WHEN sum(toneladas) FILTER (WHERE es_mineral) = 760 THEN 'PASA total_mineral=760'
         ELSE 'FALLA: ' || sum(toneladas) FILTER (WHERE es_mineral)
    END AS chk_total
FROM reportes.v_produccion_acarreo
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 17 — v_produccion_rezagado (filas coherentes, no vacía)'
\echo '  Seed: pr1 MINERAL=150 ciclos=1'
\echo '============================================================'

SELECT
    fecha, turno, tipo_material, ciclos, toneladas,
    CASE WHEN toneladas IS NULL THEN 'FALLA NULL toneladas'
         WHEN ciclos    IS NULL THEN 'FALLA NULL ciclos'
         WHEN toneladas = 150   THEN 'PASA toneladas=150'
         ELSE 'FALLA toneladas=' || toneladas
    END AS chk
FROM reportes.v_produccion_rezagado
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 18 — v_horas_equipo (base de disponibilidad)'
\echo '  EQ-001 pa1: 5428-5420.5=7.5h (2026-06-01)'
\echo '  EQ-001 pa2: 5435-5428=7.0h   (2026-06-02)'
\echo '  EQ-002 pr1: 1207.5-1200=7.5h (2026-06-01)'
\echo '============================================================'

SELECT
    id_equipo,
    fecha,
    horas_operativas,
    CASE WHEN horas_operativas IS NULL THEN 'FALLA NULL horas_operativas'
         WHEN horas_operativas <= 0    THEN 'FALLA horas<=0'
         ELSE 'PASA'
    END AS chk
FROM reportes.v_horas_equipo
ORDER BY id_equipo, fecha;

-- Verificar valores puntuales de horas
SELECT
    e.codigo AS equipo,
    vh.fecha,
    vh.horas_operativas,
    CASE
        WHEN e.codigo='EQ-001' AND vh.fecha='2026-06-01' AND vh.horas_operativas=7.5 THEN 'PASA EQ-001/06-01=7.5h'
        WHEN e.codigo='EQ-001' AND vh.fecha='2026-06-02' AND vh.horas_operativas=7.0 THEN 'PASA EQ-001/06-02=7.0h'
        WHEN e.codigo='EQ-002' AND vh.fecha='2026-06-01' AND vh.horas_operativas=7.5 THEN 'PASA EQ-002/06-01=7.5h'
        ELSE 'INFO equipo=' || e.codigo || ' fecha=' || vh.fecha || ' horas=' || vh.horas_operativas
    END AS chk
FROM reportes.v_horas_equipo vh
JOIN catalogos.equipo e ON e.id = vh.id_equipo
ORDER BY e.codigo, vh.fecha;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 19 — v_productividad_barrenacion (coherencia)'
\echo '  pb1: horas_percusion=2108.5-2100=8.5h, metros=20'
\echo '  metros_por_hora = 20/8.5 ≈ 2.35'
\echo '============================================================'

SELECT
    equipo,
    tipo_barrenacion,
    fecha,
    metros,
    horas_percusion,
    metros_por_hora,
    CASE WHEN metros          IS NULL THEN 'FALLA NULL metros'
         WHEN horas_percusion IS NULL THEN 'FALLA NULL horas_percusion'
         WHEN metros_por_hora IS NULL AND horas_percusion > 0
              THEN 'FALLA NULL metros_por_hora con horas>0 (NULLIF roto?)'
         WHEN metros = 20 THEN 'PASA metros=20'
         ELSE 'FALLA metros=' || metros
    END AS chk_metros,
    CASE WHEN metros_por_hora = round(20.0/8.5, 2)
         THEN 'PASA metros/h=' || metros_por_hora
         ELSE 'INFO metros/h=' || coalesce(metros_por_hora::text,'NULL') || ' esperado≈' || round(20.0/8.5,2)
    END AS chk_ratio
FROM reportes.v_productividad_barrenacion
WHERE id_mina = :'m_cien'::uuid;


\echo ''
\echo '============================================================'
\echo 'BLOQUE 20 — CASOS BORDE GLOBALES'
\echo '============================================================'

-- 20-A: Periodos/minas sin datos en mv_plan_vs_real_mensual
--       Plan ANUAL 2026 no tiene real => diferencia=-12000, apego NULL
SELECT
    periodo,
    ton_plan,
    ton_real,
    diferencia,
    apego_pct,
    CASE WHEN periodo='2026' AND ton_plan=12000 AND ton_real=0 AND diferencia=-12000 AND apego_pct IS NULL
         THEN 'PASA borde: ANUAL sin real (apego=NULL)'
         WHEN periodo='2026'
         THEN 'FALLA borde ANUAL: plan=' || ton_plan || ' real=' || ton_real ||
              ' dif=' || diferencia || ' apego=' || coalesce(apego_pct::text,'NULL')
         ELSE 'INFO periodo=' || periodo
    END AS chk_borde
FROM reportes.mv_plan_vs_real_mensual
WHERE id_mina = :'m_cien'::uuid
ORDER BY periodo;

-- 20-B: v_balance_cargas — minas sin partes de acarreo no aparecen (vista filtra por existencia)
SELECT
    mina,
    periodo_mes,
    CASE WHEN ton_total = 0 AND ton_mineral = 0 AND ton_tepetate = 0
         THEN 'INFO: mina con solo ceros (posible si hay partes vacíos)'
         ELSE 'PASA: mina ' || mina || ' tiene datos'
    END AS chk
FROM reportes.v_balance_cargas;

-- 20-C: v_reconciliacion — ninguna diferencia debería dar NULL cuando ambas fuentes existen
SELECT
    segmento,
    dif_tumbe_reserva,
    dif_planta_tumbe,
    CASE WHEN dif_tumbe_reserva IS NULL AND
              (SELECT count(*) FROM reconciliacion.medicion med
               JOIN reconciliacion.segmento s ON s.id = med.id_segmento
               WHERE s.codigo = segmento AND med.fuente IN ('TUMBE','RESERVA')) = 2
         THEN 'FALLA: ambas fuentes existen pero diferencia es NULL'
         ELSE 'PASA: diferencias calculadas correctamente'
    END AS chk_nulls_reconciliacion
FROM reportes.v_reconciliacion;

-- 20-D: Todos los apego_pct en mv_plan_vs_real_mensual son <= 200 (sanity)
SELECT
    count(*) AS filas_apego_sospechoso
FROM reportes.mv_plan_vs_real_mensual
WHERE apego_pct > 200 OR apego_pct < 0;

-- 20-E: Todas las utilizacion_pct en v_disponibilidad_equipo están en [0,100]
SELECT
    count(*) AS filas_util_fuera_rango
FROM reportes.v_disponibilidad_equipo
WHERE utilizacion_pct > 100 OR utilizacion_pct < 0;

-- 20-F: v_inversion_obra no tiene inversion negativa
SELECT
    count(*) AS inversiones_negativas
FROM reportes.v_inversion_obra
WHERE inversion < 0;


\echo ''
\echo '============================================================'
\echo 'RESUMEN FINAL — Conteo de vistas con datos'
\echo '============================================================'

SELECT 'v_balance_cargas'          AS vista, count(*) AS filas FROM reportes.v_balance_cargas
UNION ALL
SELECT 'mv_plan_vs_real_mensual',           count(*) FROM reportes.mv_plan_vs_real_mensual
UNION ALL
SELECT 'v_avance_barrenacion',              count(*) FROM reportes.v_avance_barrenacion
UNION ALL
SELECT 'v_reconciliacion',                  count(*) FROM reportes.v_reconciliacion
UNION ALL
SELECT 'v_disponibilidad_equipo',           count(*) FROM reportes.v_disponibilidad_equipo
UNION ALL
SELECT 'v_tiempos_vs_estandar',             count(*) FROM reportes.v_tiempos_vs_estandar
UNION ALL
SELECT 'v_produccion_vs_molienda',          count(*) FROM reportes.v_produccion_vs_molienda
UNION ALL
SELECT 'v_balance_metalurgico',             count(*) FROM reportes.v_balance_metalurgico
UNION ALL
SELECT 'v_obras_prioritarias',              count(*) FROM reportes.v_obras_prioritarias
UNION ALL
SELECT 'v_inversion_obra',                  count(*) FROM reportes.v_inversion_obra
UNION ALL
SELECT 'v_presupuesto_vs_real',             count(*) FROM reportes.v_presupuesto_vs_real
UNION ALL
SELECT 'v_costo_unitario',                  count(*) FROM reportes.v_costo_unitario
UNION ALL
SELECT 'v_costo_estimacion',                count(*) FROM reportes.v_costo_estimacion
UNION ALL
SELECT 'v_consumo_explosivo',               count(*) FROM reportes.v_consumo_explosivo
UNION ALL
SELECT 'v_plan_periodo',                    count(*) FROM reportes.v_plan_periodo
UNION ALL
SELECT 'v_produccion_acarreo',              count(*) FROM reportes.v_produccion_acarreo
UNION ALL
SELECT 'v_produccion_rezagado',             count(*) FROM reportes.v_produccion_rezagado
UNION ALL
SELECT 'v_horas_equipo',                    count(*) FROM reportes.v_horas_equipo
UNION ALL
SELECT 'v_productividad_barrenacion',       count(*) FROM reportes.v_productividad_barrenacion
ORDER BY vista;
