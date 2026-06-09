-- ============================================================
-- 04 · Planeación — plan maestro de bloques / rebajes
-- ============================================================
-- La OTRA columna vertebral (el "Plan"). Bloques programados con
-- dimensiones, dilución, leyes diluidas, NSR, cut-off y metas por
-- periodo en los 6 horizontes. Unicidad por índice PARCIAL.

CREATE TABLE planeacion.rebaje (              -- stope físico de la mina
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID,                               -- liga opcional a la labor (catalogos.obra)
  codigo TEXT NOT NULL,                       -- p.ej. TASE2040E
  descripcion TEXT,
  id_tipo_obra UUID,
  id_sistema_minado UUID,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_rebaje_estado CHECK (estado IN ('ACTIVO','INACTIVO','AGOTADO')),
  CONSTRAINT fk_rebaje_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_rebaje_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_rebaje_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_rebaje_tipo_obra FOREIGN KEY (id_tipo_obra) REFERENCES catalogos.tipo_obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_rebaje_sistema_minado FOREIGN KEY (id_sistema_minado) REFERENCES catalogos.sistema_minado(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_rebaje ON planeacion.rebaje (id_empresa, id_mina, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE planeacion.plan (                -- escenario de plan (p.ej. CUTS2026)
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  nombre TEXT NOT NULL,
  anio SMALLINT NOT NULL,
  t_inversion TEXT NOT NULL DEFAULT 'OPEX',   -- OPEX / CAPEX
  enfoque TEXT NOT NULL DEFAULT 'BASE',       -- BASE / OPERATIVO / BUDGET / FORECAST
  descripcion TEXT,
  estado TEXT NOT NULL DEFAULT 'VIGENTE',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_plan_tinv CHECK (t_inversion IN ('OPEX','CAPEX')),
  CONSTRAINT ck_plan_enfoque CHECK (enfoque IN ('BASE','OPERATIVO','BUDGET','FORECAST')),
  CONSTRAINT ck_plan_estado CHECK (estado IN ('VIGENTE','CERRADO','BORRADOR')),
  CONSTRAINT fk_plan_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_plan_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_plan ON planeacion.plan (id_empresa, id_mina, nombre, anio, enfoque) WHERE eliminado_en IS NULL;

CREATE TABLE planeacion.bloque_programado (   -- una fila de la lámina maestra de plan
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_plan UUID NOT NULL,
  id_rebaje UUID NOT NULL,
  codigo_bloque TEXT NOT NULL,                -- Block, p.ej. 819
  fecha_inicio DATE,                          -- Start
  largo NUMERIC(8,2),
  ancho_veta NUMERIC(8,2),                    -- AnchoV
  alto NUMERIC(8,2),
  fraccion NUMERIC(6,3),                      -- Fracc
  waste NUMERIC(6,3),
  minado NUMERIC(6,3),
  factor_dilucion NUMERIC(7,4),              -- FacDilucion
  dilucion_pct NUMERIC(6,2),                  -- Dilucion (%)
  toneladas NUMERIC(14,3),                    -- Ton
  toneladas_diluidas NUMERIC(14,3),          -- Ton Dil
  ley_au NUMERIC(10,4),                       -- AUd
  ley_ag NUMERIC(10,4),                       -- AGd
  ley_pb NUMERIC(10,4),                       -- PBd
  ley_zn NUMERIC(10,4),                       -- ZNd
  nsr NUMERIC(14,4),
  clasif_h TEXT,                              -- CLASIFH
  corte_h NUMERIC(14,4),                      -- CorteH (cut-off del bloque)
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT fk_bloque_prog_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bloque_prog_plan FOREIGN KEY (id_plan) REFERENCES planeacion.plan(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bloque_prog_rebaje FOREIGN KEY (id_rebaje) REFERENCES planeacion.rebaje(id) ON DELETE RESTRICT
);
CREATE INDEX ix_bloque_prog_plan ON planeacion.bloque_programado (id_plan);
CREATE INDEX ix_bloque_prog_rebaje ON planeacion.bloque_programado (id_rebaje);
CREATE UNIQUE INDEX uq_bloque_prog ON planeacion.bloque_programado (id_empresa, id_plan, id_rebaje, codigo_bloque) WHERE eliminado_en IS NULL;

-- Meta del bloque por periodo/horizonte (PANUAL..PDIARIO del ER) normalizado.
CREATE TABLE planeacion.meta_periodo (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_bloque_programado UUID NOT NULL,
  horizonte TEXT NOT NULL,                    -- ANUAL/SEMESTRAL/TRIMESTRAL/MENSUAL/SEMANAL/DIARIO
  periodo_etiqueta TEXT NOT NULL,             -- 2026 / 2026-S1 / 2026-T3 / 2026-09 / 2026-W36 / 2026-09-15
  toneladas_plan NUMERIC(14,3),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_meta_horizonte CHECK (horizonte IN ('ANUAL','SEMESTRAL','TRIMESTRAL','MENSUAL','SEMANAL','DIARIO')),
  CONSTRAINT fk_meta_periodo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_meta_periodo_bloque FOREIGN KEY (id_bloque_programado) REFERENCES planeacion.bloque_programado(id) ON DELETE RESTRICT
);
CREATE INDEX ix_meta_periodo_bloque ON planeacion.meta_periodo (id_bloque_programado);
CREATE UNIQUE INDEX uq_meta_periodo ON planeacion.meta_periodo (id_empresa, id_bloque_programado, horizonte, periodo_etiqueta) WHERE eliminado_en IS NULL;
