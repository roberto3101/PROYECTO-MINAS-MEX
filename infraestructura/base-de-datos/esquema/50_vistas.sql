-- ============================================================
-- 50 · Vistas base de entregables (capa reportes)
-- ============================================================
-- Cada familia de los "+2,000 entregables" es una parametrización
-- (por periodo / mina / obra / equipo) de estas vistas base.
-- Se carga AL FINAL: depende de todas las tablas de todas las capacidades.

-- ---------- Producción (Real) ----------
CREATE VIEW reportes.v_produccion_acarreo AS
SELECT pa.id_empresa, pa.id_mina, m.nombre AS mina, pa.fecha, pa.turno,
       tm.codigo AS tipo_material, tm.es_mineral,
       count(av.id) AS viajes, coalesce(sum(av.toneladas),0) AS toneladas
FROM produccion.parte_acarreo pa
JOIN catalogos.mina m ON m.id = pa.id_mina
JOIN produccion.acarreo_viaje av ON av.id_parte_acarreo = pa.id
JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
WHERE pa.eliminado_en IS NULL
GROUP BY pa.id_empresa, pa.id_mina, m.nombre, pa.fecha, pa.turno, tm.codigo, tm.es_mineral;

CREATE VIEW reportes.v_produccion_rezagado AS
SELECT pr.id_empresa, pr.id_mina, m.nombre AS mina, pr.fecha, pr.turno,
       tm.codigo AS tipo_material, tm.es_mineral,
       count(rc.id) AS ciclos, coalesce(sum(rc.toneladas),0) AS toneladas
FROM produccion.parte_rezagado pr
JOIN catalogos.mina m ON m.id = pr.id_mina
JOIN produccion.rezagado_ciclo rc ON rc.id_parte_rezagado = pr.id
JOIN catalogos.tipo_mineral tm ON tm.id = rc.id_tipo_mineral
WHERE pr.eliminado_en IS NULL
GROUP BY pr.id_empresa, pr.id_mina, m.nombre, pr.fecha, pr.turno, tm.codigo, tm.es_mineral;

-- Balance de cargas: mineral vs tepetate por mina/mes
CREATE VIEW reportes.v_balance_cargas AS
SELECT id_empresa, id_mina, mina, to_char(fecha,'YYYY-MM') AS periodo_mes,
       coalesce(sum(toneladas) FILTER (WHERE es_mineral), 0)     AS ton_mineral,
       coalesce(sum(toneladas) FILTER (WHERE NOT es_mineral), 0) AS ton_tepetate,
       coalesce(sum(toneladas), 0)                               AS ton_total
FROM reportes.v_produccion_acarreo
GROUP BY id_empresa, id_mina, mina, to_char(fecha,'YYYY-MM');

-- Avance de barrenación por obra
CREATE VIEW reportes.v_avance_barrenacion AS
SELECT pb.id_empresa, pb.id_mina, m.nombre AS mina, o.codigo AS obra,
       pb.tipo_barrenacion, ba.actividad, pb.fecha,
       count(ba.id) AS registros,
       coalesce(sum(ba.longitud),0) AS metros, coalesce(sum(ba.no_barrenos),0) AS barrenos
FROM produccion.parte_barrenacion pb
JOIN catalogos.mina m ON m.id = pb.id_mina
JOIN catalogos.obra o ON o.id = pb.id_obra
JOIN produccion.barrenacion_avance ba ON ba.id_parte_barrenacion = pb.id
WHERE pb.eliminado_en IS NULL
GROUP BY pb.id_empresa, pb.id_mina, m.nombre, o.codigo, pb.tipo_barrenacion, ba.actividad, pb.fecha;

-- ---------- Tiempos operativos (estándares de tiempos y movimientos) ----------
CREATE VIEW reportes.v_horas_equipo AS
SELECT id_empresa, id_mina, id_equipo, fecha, (horometro_final - horometro_inicial) AS horas_operativas
FROM produccion.parte_acarreo  WHERE eliminado_en IS NULL
UNION ALL
SELECT id_empresa, id_mina, id_equipo, fecha, (horometro_final - horometro_inicial)
FROM produccion.parte_rezagado WHERE eliminado_en IS NULL;

CREATE VIEW reportes.v_disponibilidad_equipo AS
WITH horas AS (
  SELECT id_empresa, id_mina, id_equipo, fecha, sum(horas_operativas) AS horas_operativas
  FROM reportes.v_horas_equipo GROUP BY id_empresa, id_mina, id_equipo, fecha
),
dem AS (
  SELECT id_empresa, id_equipo, fecha, sum(minutos)/60.0 AS horas_demora
  FROM produccion.demora_equipo GROUP BY id_empresa, id_equipo, fecha
)
SELECT h.id_empresa, h.id_mina, e.codigo AS equipo, h.id_equipo, h.fecha,
       round(h.horas_operativas, 2) AS horas_operativas,
       round(coalesce(d.horas_demora, 0), 2) AS horas_demora,
       round(h.horas_operativas / NULLIF(h.horas_operativas + coalesce(d.horas_demora,0), 0) * 100, 2) AS utilizacion_pct
FROM horas h
JOIN catalogos.equipo e ON e.id = h.id_equipo
LEFT JOIN dem d ON d.id_empresa = h.id_empresa AND d.id_equipo = h.id_equipo AND d.fecha = h.fecha;

-- Tiempos operativos REAL vs ESTÁNDAR (por tipo de equipo)
CREATE VIEW reportes.v_tiempos_vs_estandar AS
SELECT d.id_empresa, d.id_mina, d.equipo, te.codigo AS tipo_equipo, d.fecha,
       d.horas_operativas, d.horas_demora, d.utilizacion_pct,
       et.utilizacion_meta_pct,
       round(d.utilizacion_pct - et.utilizacion_meta_pct, 2) AS desvio_utilizacion_pct
FROM reportes.v_disponibilidad_equipo d
JOIN catalogos.equipo e ON e.id = d.id_equipo
JOIN catalogos.tipo_equipo te ON te.id = e.id_tipo_equipo
LEFT JOIN LATERAL (
  SELECT utilizacion_meta_pct FROM estandares.estandar_tiempo s
  WHERE s.id_empresa = d.id_empresa AND s.id_tipo_equipo = e.id_tipo_equipo AND s.eliminado_en IS NULL
  ORDER BY s.vigente_desde DESC LIMIT 1
) et ON true;

-- Productividad de barrenación (metros / hora de percusión)
-- Las horas se agregan A NIVEL PARTE (CTE horas) y los metros sumando los avances
-- (CTE avance) por separado, para evitar el fan-out del horómetro al hacer el join
-- cabecera×detalle (un parte con N avances NO debe multiplicar sus horas por N).
CREATE VIEW reportes.v_productividad_barrenacion AS
WITH horas AS (
  SELECT pb.id_empresa, pb.id_mina, pb.id_equipo, pb.tipo_barrenacion, pb.fecha,
         sum(pb.horometro_percusion_final - pb.horometro_percusion_inicial) AS horas_percusion
  FROM produccion.parte_barrenacion pb
  WHERE pb.eliminado_en IS NULL
  GROUP BY pb.id_empresa, pb.id_mina, pb.id_equipo, pb.tipo_barrenacion, pb.fecha
),
avance AS (
  SELECT pb.id_empresa, pb.id_mina, pb.id_equipo, pb.tipo_barrenacion, pb.fecha,
         sum(ba.longitud) AS metros
  FROM produccion.parte_barrenacion pb
  JOIN produccion.barrenacion_avance ba ON ba.id_parte_barrenacion = pb.id
  WHERE pb.eliminado_en IS NULL
  GROUP BY pb.id_empresa, pb.id_mina, pb.id_equipo, pb.tipo_barrenacion, pb.fecha
)
SELECT h.id_empresa, h.id_mina, e.codigo AS equipo, h.tipo_barrenacion, h.fecha,
       round(coalesce(a.metros,0), 2) AS metros,
       round(h.horas_percusion, 2) AS horas_percusion,
       round(coalesce(a.metros,0) / NULLIF(h.horas_percusion,0), 2) AS metros_por_hora
FROM horas h
JOIN catalogos.equipo e ON e.id = h.id_equipo
LEFT JOIN avance a ON a.id_empresa=h.id_empresa AND a.id_mina=h.id_mina AND a.id_equipo=h.id_equipo
                  AND a.tipo_barrenacion=h.tipo_barrenacion AND a.fecha=h.fecha;

-- ---------- Plan ----------
CREATE VIEW reportes.v_plan_periodo AS
SELECT p.id_empresa, p.id_mina, m.nombre AS mina, mp.horizonte, mp.periodo_etiqueta,
       r.codigo AS rebaje, coalesce(sum(mp.toneladas_plan),0) AS toneladas_plan
FROM planeacion.meta_periodo mp
JOIN planeacion.bloque_programado bp ON bp.id = mp.id_bloque_programado
JOIN planeacion.plan p ON p.id = bp.id_plan
JOIN planeacion.rebaje r ON r.id = bp.id_rebaje
JOIN catalogos.mina m ON m.id = p.id_mina
WHERE bp.eliminado_en IS NULL AND mp.eliminado_en IS NULL
GROUP BY p.id_empresa, p.id_mina, m.nombre, mp.horizonte, mp.periodo_etiqueta, r.codigo;

-- Plan vs Real mensual (materializada: corazón de los entregables)
CREATE MATERIALIZED VIEW reportes.mv_plan_vs_real_mensual AS
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

-- Seguimiento de obras prioritarias
CREATE VIEW reportes.v_obras_prioritarias AS
SELECT o.id_empresa, o.id_mina, m.nombre AS mina, o.codigo AS obra, o.nombre,
       coalesce((SELECT sum(av.toneladas) FROM produccion.parte_acarreo pa
                 JOIN produccion.acarreo_viaje av ON av.id_parte_acarreo = pa.id
                 JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
                 WHERE pa.id_obra = o.id AND tm.es_mineral AND pa.eliminado_en IS NULL), 0) AS ton_mineral,
       coalesce((SELECT sum(ba.longitud) FROM produccion.parte_barrenacion pb
                 JOIN produccion.barrenacion_avance ba ON ba.id_parte_barrenacion = pb.id
                 WHERE pb.id_obra = o.id AND pb.eliminado_en IS NULL), 0) AS metros_barrenados
FROM catalogos.obra o
JOIN catalogos.mina m ON m.id = o.id_mina
WHERE o.es_prioritaria AND o.eliminado_en IS NULL;

-- ---------- Reconciliación ----------
CREATE VIEW reportes.v_reconciliacion AS
SELECT s.id_empresa, s.id_mina, m.nombre AS mina, s.codigo AS segmento,
       sum(med.toneladas) FILTER (WHERE med.fuente='RESERVA')    AS ton_reserva,
       sum(med.toneladas) FILTER (WHERE med.fuente='TUMBE')      AS ton_tumbe,
       sum(med.toneladas) FILTER (WHERE med.fuente='ESTIMACION') AS ton_estimacion,
       sum(med.toneladas) FILTER (WHERE med.fuente='PLANTA')     AS ton_planta,
       sum(med.toneladas) FILTER (WHERE med.fuente='TUMBE')
         - sum(med.toneladas) FILTER (WHERE med.fuente='RESERVA') AS dif_tumbe_reserva,
       sum(med.toneladas) FILTER (WHERE med.fuente='PLANTA')
         - sum(med.toneladas) FILTER (WHERE med.fuente='TUMBE')   AS dif_planta_tumbe
FROM reconciliacion.segmento s
JOIN catalogos.mina m ON m.id = s.id_mina
LEFT JOIN reconciliacion.medicion med ON med.id_segmento = s.id
WHERE s.eliminado_en IS NULL
GROUP BY s.id_empresa, s.id_mina, m.nombre, s.codigo;

-- ---------- Beneficio: producción vs molienda + balance metalúrgico ----------
CREATE VIEW reportes.v_produccion_vs_molienda AS
WITH acarreado AS (
  SELECT pa.id_empresa, pa.id_mina, to_char(pa.fecha,'YYYY-MM') AS periodo, sum(av.toneladas) AS ton_acarreada
  FROM produccion.parte_acarreo pa
  JOIN produccion.acarreo_viaje av ON av.id_parte_acarreo = pa.id
  JOIN catalogos.tipo_mineral tm ON tm.id = av.id_tipo_mineral
  WHERE tm.es_mineral AND pa.eliminado_en IS NULL
  GROUP BY pa.id_empresa, pa.id_mina, to_char(pa.fecha,'YYYY-MM')
),
molido AS (
  SELECT id_empresa, id_mina, to_char(fecha,'YYYY-MM') AS periodo, sum(toneladas_molidas) AS ton_molida
  FROM beneficio.lote_molienda WHERE eliminado_en IS NULL
  GROUP BY id_empresa, id_mina, to_char(fecha,'YYYY-MM')
)
SELECT coalesce(a.id_empresa,mo.id_empresa) AS id_empresa,
       coalesce(a.id_mina,mo.id_mina)       AS id_mina,
       coalesce(a.periodo,mo.periodo)       AS periodo,
       coalesce(a.ton_acarreada,0)          AS ton_acarreada,
       coalesce(mo.ton_molida,0)            AS ton_molida,
       coalesce(mo.ton_molida,0) - coalesce(a.ton_acarreada,0) AS diferencia
FROM acarreado a
FULL OUTER JOIN molido mo ON a.id_empresa=mo.id_empresa AND a.id_mina=mo.id_mina AND a.periodo=mo.periodo;

-- Leyes y recuperación se pre-agregan a 1 fila por lote en subconsultas
-- separadas para evitar el producto cartesiano ley×recuperación al unir al lote.
CREATE VIEW reportes.v_balance_metalurgico AS
SELECT lm.id_empresa, lm.id_mina, m.nombre AS mina, lm.codigo_lote, lm.fecha, lm.toneladas_molidas,
       l.au_cabeza, l.au_concentrado, l.au_cola, l.ag_cabeza,
       r.rec_au_pct, r.rec_ag_pct
FROM beneficio.lote_molienda lm
JOIN catalogos.mina m ON m.id = lm.id_mina
LEFT JOIN (
  SELECT id_lote_molienda,
         max(ley_au) FILTER (WHERE punto='CABEZA')      AS au_cabeza,
         max(ley_au) FILTER (WHERE punto='CONCENTRADO') AS au_concentrado,
         max(ley_au) FILTER (WHERE punto='COLA')        AS au_cola,
         max(ley_ag) FILTER (WHERE punto='CABEZA')      AS ag_cabeza
  FROM beneficio.ley_metalurgica GROUP BY id_lote_molienda
) l ON l.id_lote_molienda = lm.id
LEFT JOIN (
  SELECT id_lote_molienda,
         max(recuperacion_pct) FILTER (WHERE metal='AU') AS rec_au_pct,
         max(recuperacion_pct) FILTER (WHERE metal='AG') AS rec_ag_pct
  FROM beneficio.recuperacion GROUP BY id_lote_molienda
) r ON r.id_lote_molienda = lm.id
WHERE lm.eliminado_en IS NULL;

-- ---------- Costos ----------
CREATE VIEW reportes.v_costo_estimacion AS
SELECT e.id_empresa, e.id_mina, m.nombre AS mina, e.periodo, e.tipo,
       coalesce(sum(ec.importe),0) AS importe
FROM costos.estimacion e
JOIN catalogos.mina m ON m.id = e.id_mina
LEFT JOIN costos.estimacion_concepto ec ON ec.id_estimacion = e.id
WHERE e.eliminado_en IS NULL
GROUP BY e.id_empresa, e.id_mina, m.nombre, e.periodo, e.tipo;

CREATE VIEW reportes.v_presupuesto_vs_real AS
WITH pres AS (
  SELECT id_empresa, id_mina, coalesce(periodo, anio::text) AS periodo, sum(monto_presupuestado) AS presupuesto
  FROM costos.presupuesto WHERE eliminado_en IS NULL GROUP BY 1,2,3
),
ejecutado AS (
  SELECT e.id_empresa, e.id_mina, e.periodo, sum(ec.importe) AS real_importe
  FROM costos.estimacion e
  JOIN costos.estimacion_concepto ec ON ec.id_estimacion = e.id
  WHERE e.eliminado_en IS NULL GROUP BY 1,2,3
)
SELECT coalesce(p.id_empresa, r.id_empresa) AS id_empresa,
       coalesce(p.id_mina, r.id_mina)       AS id_mina,
       coalesce(p.periodo, r.periodo)       AS periodo,
       coalesce(p.presupuesto, 0)           AS presupuesto,
       coalesce(r.real_importe, 0)          AS real_importe,
       coalesce(r.real_importe,0) - coalesce(p.presupuesto,0) AS variacion
FROM pres p
FULL OUTER JOIN ejecutado r
  ON p.id_empresa=r.id_empresa AND p.id_mina=r.id_mina AND p.periodo=r.periodo;

CREATE VIEW reportes.v_costo_unitario AS
SELECT cu.id_empresa, cu.id_mina, m.nombre AS mina, cu.periodo, cu.actividad, cu.area, cu.unidad, cu.valor, cu.moneda
FROM costos.costo_unitario cu
JOIN catalogos.mina m ON m.id = cu.id_mina
WHERE cu.eliminado_en IS NULL;

-- ---------- Explosivo ----------
CREATE VIEW reportes.v_consumo_explosivo AS
SELECT ce.id_empresa, ce.id_mina, m.nombre AS mina, o.codigo AS obra, ce.destino, te.familia,
       to_char(ce.fecha,'YYYY-MM') AS periodo, coalesce(sum(ce.cantidad),0) AS cantidad
FROM produccion.consumo_explosivo ce
JOIN catalogos.mina m ON m.id = ce.id_mina
JOIN catalogos.obra o ON o.id = ce.id_obra
JOIN catalogos.tipo_explosivo te ON te.id = ce.id_tipo_explosivo
GROUP BY ce.id_empresa, ce.id_mina, m.nombre, o.codigo, ce.destino, te.familia, to_char(ce.fecha,'YYYY-MM');

-- ---------- Inversiones ----------
CREATE VIEW reportes.v_inversion_obra AS
SELECT i.id_empresa, i.id_mina, m.nombre AS mina, o.codigo AS obra, i.periodo,
       coalesce(sum(i.monto),0) AS inversion, i.moneda
FROM inversiones.inversion i
JOIN catalogos.mina m ON m.id = i.id_mina
LEFT JOIN catalogos.obra o ON o.id = i.id_obra
WHERE i.eliminado_en IS NULL
GROUP BY i.id_empresa, i.id_mina, m.nombre, o.codigo, i.periodo, i.moneda;
