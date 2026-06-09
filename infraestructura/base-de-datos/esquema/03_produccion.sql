-- ============================================================
-- 03 · Producción — captura del "Real" (offline-first)
-- ============================================================
-- Patrón: cabecera (parte_*) = Raíz de Agregado (editable, borrado lógico)
--         detalle (viaje/ciclo/avance) = Evento append-only (solo creado_*).
-- La obra es entidad (catalogos.obra), referenciada por id_obra.

-- ---------- ACARREO ----------
CREATE TABLE produccion.parte_acarreo (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID NOT NULL,
  id_equipo UUID NOT NULL,              -- camión / locomotora
  id_operador UUID NOT NULL,
  id_supervisor UUID,
  fecha DATE NOT NULL,
  turno TEXT NOT NULL,                  -- M / V / N
  horometro_inicial NUMERIC(10,2) NOT NULL,
  horometro_final NUMERIC(10,2) NOT NULL,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_parte_acarreo_turno CHECK (turno IN ('M','V','N')),
  CONSTRAINT ck_parte_acarreo_horometro CHECK (horometro_final >= horometro_inicial),
  CONSTRAINT fk_parte_acarreo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_acarreo_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_acarreo_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_acarreo_equipo FOREIGN KEY (id_equipo) REFERENCES catalogos.equipo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_acarreo_operador FOREIGN KEY (id_operador) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_acarreo_supervisor FOREIGN KEY (id_supervisor) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT
);
CREATE INDEX ix_parte_acarreo_mina_fecha ON produccion.parte_acarreo (id_empresa, id_mina, fecha);
CREATE INDEX ix_parte_acarreo_obra ON produccion.parte_acarreo (id_obra);

CREATE TABLE produccion.acarreo_viaje (       -- evento: cada viaje del camión
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_parte_acarreo UUID NOT NULL,
  desde TEXT NOT NULL,
  hasta TEXT NOT NULL,
  id_tipo_mineral UUID NOT NULL,       -- mineral vs tepetate
  toneladas NUMERIC(12,3),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT fk_acarreo_viaje_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_acarreo_viaje_parte FOREIGN KEY (id_parte_acarreo) REFERENCES produccion.parte_acarreo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_acarreo_viaje_tipo_mineral FOREIGN KEY (id_tipo_mineral) REFERENCES catalogos.tipo_mineral(id) ON DELETE RESTRICT
);
CREATE INDEX ix_acarreo_viaje_parte ON produccion.acarreo_viaje (id_parte_acarreo);

-- ---------- REZAGADO (scooptram) ----------
CREATE TABLE produccion.parte_rezagado (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID NOT NULL,
  id_equipo UUID NOT NULL,             -- scooptram
  id_operador UUID NOT NULL,
  id_supervisor UUID,
  fecha DATE NOT NULL,
  turno TEXT NOT NULL,
  horometro_inicial NUMERIC(10,2) NOT NULL,
  horometro_final NUMERIC(10,2) NOT NULL,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_parte_rezagado_turno CHECK (turno IN ('M','V','N')),
  CONSTRAINT ck_parte_rezagado_horometro CHECK (horometro_final >= horometro_inicial),
  CONSTRAINT fk_parte_rezagado_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_rezagado_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_rezagado_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_rezagado_equipo FOREIGN KEY (id_equipo) REFERENCES catalogos.equipo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_rezagado_operador FOREIGN KEY (id_operador) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_rezagado_supervisor FOREIGN KEY (id_supervisor) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT
);
CREATE INDEX ix_parte_rezagado_mina_fecha ON produccion.parte_rezagado (id_empresa, id_mina, fecha);
CREATE INDEX ix_parte_rezagado_obra ON produccion.parte_rezagado (id_obra);

CREATE TABLE produccion.rezagado_ciclo (      -- evento: cada ciclo de carga del scooptram
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_parte_rezagado UUID NOT NULL,
  desde TEXT NOT NULL,
  hasta TEXT NOT NULL,
  id_tipo_mineral UUID NOT NULL,
  toneladas NUMERIC(12,3),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT fk_rezagado_ciclo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_rezagado_ciclo_parte FOREIGN KEY (id_parte_rezagado) REFERENCES produccion.parte_rezagado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_rezagado_ciclo_tipo_mineral FOREIGN KEY (id_tipo_mineral) REFERENCES catalogos.tipo_mineral(id) ON DELETE RESTRICT
);
CREATE INDEX ix_rezagado_ciclo_parte ON produccion.rezagado_ciclo (id_parte_rezagado);

-- ---------- BARRENACIÓN (unifica 4 tablas del ER con discriminador) ----------
CREATE TABLE produccion.parte_barrenacion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  tipo_barrenacion TEXT NOT NULL,      -- LINEAL / MAQ_PIERNA / ANCLADOR / LARGA
  id_mina UUID NOT NULL,
  id_obra UUID NOT NULL,
  id_equipo UUID NOT NULL,             -- jumbo
  id_capitan_mina UUID NOT NULL,
  id_supervisor UUID,
  id_operador UUID NOT NULL,
  id_ayudante UUID,
  id_actividad UUID,
  fecha DATE NOT NULL,
  turno TEXT NOT NULL,
  horometro_diesel_inicial NUMERIC(10,2),
  horometro_diesel_final NUMERIC(10,2),
  horometro_electrico_inicial NUMERIC(10,2),
  horometro_electrico_final NUMERIC(10,2),
  horometro_percusion_inicial NUMERIC(10,2),
  horometro_percusion_final NUMERIC(10,2),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_parte_barr_turno CHECK (turno IN ('M','V','N')),
  CONSTRAINT ck_parte_barr_tipo CHECK (tipo_barrenacion IN ('LINEAL','MAQ_PIERNA','ANCLADOR','LARGA')),
  CONSTRAINT fk_parte_barr_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_equipo FOREIGN KEY (id_equipo) REFERENCES catalogos.equipo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_capitan FOREIGN KEY (id_capitan_mina) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_supervisor FOREIGN KEY (id_supervisor) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_operador FOREIGN KEY (id_operador) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_ayudante FOREIGN KEY (id_ayudante) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT,
  CONSTRAINT fk_parte_barr_actividad FOREIGN KEY (id_actividad) REFERENCES catalogos.actividad(id) ON DELETE RESTRICT
);
CREATE INDEX ix_parte_barr_mina_fecha ON produccion.parte_barrenacion (id_empresa, id_mina, fecha);
CREATE INDEX ix_parte_barr_obra ON produccion.parte_barrenacion (id_obra);
CREATE INDEX ix_parte_barr_tipo ON produccion.parte_barrenacion (id_empresa, tipo_barrenacion);

CREATE TABLE produccion.barrenacion_avance (  -- evento: avance barrenado por actividad
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_parte_barrenacion UUID NOT NULL,
  actividad TEXT NOT NULL,             -- BARRENACION / RE_BARRENACION / ESCAREO
  lugar TEXT NOT NULL,
  no_secciones SMALLINT,
  longitud NUMERIC(7,2) NOT NULL,
  no_barrenos INTEGER NOT NULL,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_barr_avance_actividad CHECK (actividad IN ('BARRENACION','RE_BARRENACION','ESCAREO')),
  CONSTRAINT ck_barr_avance_longitud CHECK (longitud >= 0),
  CONSTRAINT ck_barr_avance_barrenos CHECK (no_barrenos >= 0),
  CONSTRAINT fk_barr_avance_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_barr_avance_parte FOREIGN KEY (id_parte_barrenacion) REFERENCES produccion.parte_barrenacion(id) ON DELETE RESTRICT
);
CREATE INDEX ix_barr_avance_parte ON produccion.barrenacion_avance (id_parte_barrenacion);

-- "Ejecutado" del ER (lista fija múltiple): qué secciones de barreno se ejecutaron
-- en este avance. Relación N:M con catalogos.tipo_barreno. Append-only.
CREATE TABLE produccion.barrenacion_ejecutado (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_barrenacion_avance UUID NOT NULL,
  id_tipo_barreno UUID NOT NULL,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT uq_barr_ejecutado UNIQUE (id_barrenacion_avance, id_tipo_barreno),
  CONSTRAINT fk_barr_ejec_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_barr_ejec_avance FOREIGN KEY (id_barrenacion_avance) REFERENCES produccion.barrenacion_avance(id) ON DELETE RESTRICT,
  CONSTRAINT fk_barr_ejec_tipo FOREIGN KEY (id_tipo_barreno) REFERENCES catalogos.tipo_barreno(id) ON DELETE RESTRICT
);
CREATE INDEX ix_barr_ejec_avance ON produccion.barrenacion_ejecutado (id_barrenacion_avance);
CREATE INDEX ix_barr_ejec_tipo   ON produccion.barrenacion_ejecutado (id_tipo_barreno);

-- ---------- DEMORAS de equipo (tiempos operativos) ----------
CREATE TABLE produccion.demora_equipo (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_equipo UUID NOT NULL,
  id_tipo_demora UUID NOT NULL,
  fecha DATE NOT NULL,
  turno TEXT NOT NULL,
  minutos INTEGER NOT NULL,
  observacion TEXT,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_demora_turno CHECK (turno IN ('M','V','N')),
  CONSTRAINT ck_demora_minutos CHECK (minutos >= 0),
  CONSTRAINT fk_demora_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_demora_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_demora_equipo FOREIGN KEY (id_equipo) REFERENCES catalogos.equipo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_demora_tipo FOREIGN KEY (id_tipo_demora) REFERENCES catalogos.tipo_demora(id) ON DELETE RESTRICT
);
CREATE INDEX ix_demora_equipo_fecha ON produccion.demora_equipo (id_empresa, id_equipo, fecha);
