-- ============================================================
-- 11 · Índices de rendimiento (FKs + parciales para escala)
-- ============================================================
-- PostgreSQL NO crea índices automáticos en las columnas FK del hijo.
-- A escala (millones de filas) esto provoca seq scans en joins y en
-- DELETE/UPDATE del padre. Aquí se cubren las FKs y patrones de consulta
-- detectados en la auditoría de rendimiento. Los índices en columnas
-- opcionales son parciales (WHERE ... IS NOT NULL) para ahorrar espacio.

-- ---------- Producción: eventos de alto volumen ----------
CREATE INDEX ix_acarreo_viaje_tipo_mineral  ON produccion.acarreo_viaje (id_tipo_mineral);
CREATE INDEX ix_rezagado_ciclo_tipo_mineral ON produccion.rezagado_ciclo (id_tipo_mineral);
CREATE INDEX ix_demora_tipo_demora          ON produccion.demora_equipo (id_tipo_demora);
CREATE INDEX ix_consumo_exp_obra            ON produccion.consumo_explosivo (id_empresa, id_obra, fecha);
CREATE INDEX ix_consumo_exp_tipo            ON produccion.consumo_explosivo (id_tipo_explosivo);
CREATE INDEX ix_consumo_exp_parte           ON produccion.consumo_explosivo (id_parte_barrenacion) WHERE id_parte_barrenacion IS NOT NULL;

-- ---------- Producción: FKs de equipo/persona en las cabeceras ----------
CREATE INDEX ix_parte_acarreo_equipo     ON produccion.parte_acarreo (id_equipo);
CREATE INDEX ix_parte_acarreo_operador   ON produccion.parte_acarreo (id_operador);
CREATE INDEX ix_parte_acarreo_supervisor ON produccion.parte_acarreo (id_supervisor) WHERE id_supervisor IS NOT NULL;
CREATE INDEX ix_parte_rezagado_equipo     ON produccion.parte_rezagado (id_equipo);
CREATE INDEX ix_parte_rezagado_operador   ON produccion.parte_rezagado (id_operador);
CREATE INDEX ix_parte_rezagado_supervisor ON produccion.parte_rezagado (id_supervisor) WHERE id_supervisor IS NOT NULL;
CREATE INDEX ix_parte_barr_equipo      ON produccion.parte_barrenacion (id_equipo);
CREATE INDEX ix_parte_barr_operador    ON produccion.parte_barrenacion (id_operador);
CREATE INDEX ix_parte_barr_capitan     ON produccion.parte_barrenacion (id_capitan_mina);
CREATE INDEX ix_parte_barr_supervisor  ON produccion.parte_barrenacion (id_supervisor) WHERE id_supervisor IS NOT NULL;
CREATE INDEX ix_parte_barr_ayudante    ON produccion.parte_barrenacion (id_ayudante) WHERE id_ayudante IS NOT NULL;

-- ---------- Producción: índices parciales (solo activos) en cabeceras ----------
CREATE INDEX ix_parte_acarreo_activos  ON produccion.parte_acarreo (id_empresa, id_mina, fecha) WHERE eliminado_en IS NULL;
CREATE INDEX ix_parte_rezagado_activos ON produccion.parte_rezagado (id_empresa, id_mina, fecha) WHERE eliminado_en IS NULL;
CREATE INDEX ix_parte_barr_activos     ON produccion.parte_barrenacion (id_empresa, id_mina, fecha) WHERE eliminado_en IS NULL;

-- ---------- Catálogos: FKs de clasificación ----------
CREATE INDEX ix_equipo_tipo_equipo    ON catalogos.equipo (id_tipo_equipo);
CREATE INDEX ix_equipo_modulo_trabajo ON catalogos.equipo (id_modulo_trabajo);
CREATE INDEX ix_empleado_departamento ON catalogos.empleado (id_departamento) WHERE id_departamento IS NOT NULL;
CREATE INDEX ix_empleado_puesto       ON catalogos.empleado (id_puesto) WHERE id_puesto IS NOT NULL;
CREATE INDEX ix_empleado_actividad    ON catalogos.empleado (id_actividad) WHERE id_actividad IS NOT NULL;
CREATE INDEX ix_mineral_tipo_mineral  ON catalogos.mineral (id_tipo_mineral);
CREATE INDEX ix_obra_tipo_obra        ON catalogos.obra (id_tipo_obra) WHERE id_tipo_obra IS NOT NULL;

-- ---------- Planeación ----------
CREATE INDEX ix_rebaje_mina ON planeacion.rebaje (id_mina);
CREATE INDEX ix_rebaje_obra ON planeacion.rebaje (id_obra) WHERE id_obra IS NOT NULL;
CREATE INDEX ix_meta_periodo_horizonte ON planeacion.meta_periodo (id_empresa, horizonte, periodo_etiqueta) WHERE eliminado_en IS NULL;

-- ---------- Reconciliación ----------
CREATE INDEX ix_segmento_rebaje ON reconciliacion.segmento (id_rebaje) WHERE id_rebaje IS NOT NULL;

-- ---------- Beneficio / Estándares ----------
CREATE INDEX ix_lote_mina            ON beneficio.lote_molienda (id_mina);
CREATE INDEX ix_estandar_tiempo_tipo ON estandares.estandar_tiempo (id_tipo_equipo);
CREATE INDEX ix_estandar_prod_act    ON estandares.estandar_productividad (id_actividad);

-- ---------- Costos ----------
CREATE INDEX ix_estimacion_mina        ON costos.estimacion (id_mina);
CREATE INDEX ix_estimacion_contratista ON costos.estimacion (id_contratista) WHERE id_contratista IS NOT NULL;
CREATE INDEX ix_estimacion_concepto_activos ON costos.estimacion_concepto (id_estimacion) WHERE eliminado_en IS NULL;
CREATE INDEX ix_cutoff_rebaje          ON costos.cutoff_variable (id_empresa, id_rebaje, periodo) WHERE id_rebaje IS NOT NULL;

-- ---------- Inversiones ----------
CREATE INDEX ix_inversion_obra_activos ON inversiones.inversion (id_empresa, id_obra, periodo) WHERE eliminado_en IS NULL;
CREATE INDEX ix_acero_obra             ON inversiones.consumo_acero (id_obra) WHERE id_obra IS NOT NULL;

-- ---------- Cobertura de FK a dimensiones (joins/filtros inversos) ----------
-- Auditoría de FK: se indexan las FK a dimensiones donde un join/filtro inverso
-- aporta valor. Las FK id_mina restantes YA están cubiertas para consulta
-- multi-tenant por sus índices compuestos (id_empresa, id_mina, ...); y como se
-- usa BORRADO LÓGICO (no DELETE físico de padres), no requieren índice de
-- integridad dedicado. La FK id_empresa en eventos masivos NO se indexa a
-- propósito (baja cardinalidad → contraproducente en tablas de alto volumen).
CREATE INDEX ix_rebaje_tipo_obra      ON planeacion.rebaje (id_tipo_obra) WHERE id_tipo_obra IS NOT NULL;
CREATE INDEX ix_rebaje_sistema_minado ON planeacion.rebaje (id_sistema_minado) WHERE id_sistema_minado IS NOT NULL;
CREATE INDEX ix_parte_barr_actividad  ON produccion.parte_barrenacion (id_actividad) WHERE id_actividad IS NOT NULL;
CREATE INDEX ix_estandar_prod_tipo_eq ON estandares.estandar_productividad (id_tipo_equipo) WHERE id_tipo_equipo IS NOT NULL;
CREATE INDEX ix_demora_equipo_fk      ON produccion.demora_equipo (id_equipo);
CREATE INDEX ix_demora_mina           ON produccion.demora_equipo (id_empresa, id_mina, fecha);
CREATE INDEX ix_cutoff_rebaje_fk      ON costos.cutoff_variable (id_rebaje) WHERE id_rebaje IS NOT NULL;
CREATE INDEX ix_inversion_obra_fk     ON inversiones.inversion (id_obra) WHERE id_obra IS NOT NULL;
CREATE INDEX ix_consumo_exp_obra_fk   ON produccion.consumo_explosivo (id_obra);
CREATE INDEX ix_activo_mina           ON inversiones.activo (id_empresa, id_mina);
