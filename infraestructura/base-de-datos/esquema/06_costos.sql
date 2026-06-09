-- ============================================================
-- 06 · Costos — estimaciones, conceptos y presupuesto
-- ============================================================
-- Estimaciones (lo medido y pagado, a contratistas) con su detalle,
-- y el presupuesto para el plan vs real económico.
-- Unicidad por índice PARCIAL (solo activos).

CREATE TABLE costos.contratista (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL,
  nombre TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_contratista_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_contratista_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_contratista ON costos.contratista (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE costos.estimacion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_contratista UUID,
  numero TEXT NOT NULL,
  tipo TEXT NOT NULL DEFAULT 'DESARROLLO',    -- DESARROLLO / ACARREO / OTRO
  periodo TEXT NOT NULL,                       -- YYYY-MM
  fecha DATE NOT NULL,
  monto_total NUMERIC(16,2) NOT NULL DEFAULT 0,
  estado TEXT NOT NULL DEFAULT 'BORRADOR',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_estimacion_tipo CHECK (tipo IN ('DESARROLLO','ACARREO','OTRO')),
  CONSTRAINT ck_estimacion_estado CHECK (estado IN ('BORRADOR','APROBADA','PAGADA')),
  CONSTRAINT fk_estimacion_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_estimacion_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_estimacion_contratista FOREIGN KEY (id_contratista) REFERENCES costos.contratista(id) ON DELETE RESTRICT
);
CREATE INDEX ix_estimacion_mina_periodo ON costos.estimacion (id_empresa, id_mina, periodo);
CREATE UNIQUE INDEX uq_estimacion ON costos.estimacion (id_empresa, id_mina, numero) WHERE eliminado_en IS NULL;

CREATE TABLE costos.estimacion_concepto (     -- detalle mutable: se edita mientras la estimación es borrador
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_estimacion UUID NOT NULL,
  concepto TEXT NOT NULL,
  rubro TEXT,
  actividad TEXT,
  centro_costo TEXT,
  departamento TEXT,
  cantidad NUMERIC(14,3) NOT NULL DEFAULT 0,
  precio_unitario NUMERIC(16,4) NOT NULL DEFAULT 0,
  importe NUMERIC(16,2) NOT NULL DEFAULT 0,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_concepto_importe CHECK (importe >= 0),
  CONSTRAINT fk_concepto_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_concepto_estimacion FOREIGN KEY (id_estimacion) REFERENCES costos.estimacion(id) ON DELETE RESTRICT
);
CREATE INDEX ix_concepto_estimacion ON costos.estimacion_concepto (id_estimacion);

CREATE TABLE costos.presupuesto (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  anio SMALLINT NOT NULL,
  periodo TEXT,                               -- NULL = anual; YYYY-MM = mensual
  rubro TEXT,
  departamento TEXT,
  actividad TEXT,
  monto_presupuestado NUMERIC(16,2) NOT NULL DEFAULT 0,
  moneda TEXT NOT NULL DEFAULT 'USD',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_presupuesto_moneda CHECK (moneda IN ('USD','MXN')),
  CONSTRAINT fk_presupuesto_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_presupuesto_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT
);
CREATE INDEX ix_presupuesto_mina_anio ON costos.presupuesto (id_empresa, id_mina, anio);

CREATE TABLE costos.costo_unitario (          -- KPI: costo por unidad por actividad/área/periodo
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  periodo TEXT NOT NULL,                       -- YYYY-MM
  actividad TEXT,
  area TEXT,
  unidad TEXT NOT NULL,                         -- USD_TON / USD_METRO / USD_M3 / USD_ONZA_EQ
  valor NUMERIC(16,4) NOT NULL DEFAULT 0,
  moneda TEXT NOT NULL DEFAULT 'USD',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_costo_unitario_unidad CHECK (unidad IN ('USD_TON','USD_METRO','USD_M3','USD_ONZA_EQ')),
  CONSTRAINT ck_costo_unitario_moneda CHECK (moneda IN ('USD','MXN')),
  CONSTRAINT fk_costo_unitario_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_costo_unitario_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT
);
CREATE INDEX ix_costo_unitario_mina_periodo ON costos.costo_unitario (id_empresa, id_mina, periodo);

CREATE TABLE costos.cutoff_variable (         -- ley de corte variable por rebaje/zona/periodo
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_rebaje UUID,
  zona TEXT,
  periodo TEXT NOT NULL,                        -- YYYY-MM
  cutoff_valor NUMERIC(14,4) NOT NULL,
  unidad TEXT NOT NULL DEFAULT 'USD_TON',       -- base del cut-off
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT fk_cutoff_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_cutoff_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_cutoff_rebaje FOREIGN KEY (id_rebaje) REFERENCES planeacion.rebaje(id) ON DELETE RESTRICT
);
CREATE INDEX ix_cutoff_mina_periodo ON costos.cutoff_variable (id_empresa, id_mina, periodo);
