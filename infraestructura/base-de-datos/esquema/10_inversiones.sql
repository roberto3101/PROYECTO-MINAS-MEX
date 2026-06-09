-- ============================================================
-- 10 · Inversiones — inversión por obra/periodo, activos, acero
-- ============================================================
-- Control de inversiones del PDF: capital invertido por obra,
-- activos capitalizables y consumo de acero.

CREATE TABLE inversiones.activo (             -- activo capitalizable
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  codigo TEXT NOT NULL,
  descripcion TEXT NOT NULL,
  tipo_activo TEXT,
  costo_adquisicion NUMERIC(16,2) NOT NULL DEFAULT 0,
  fecha_adquisicion DATE,
  vida_util_meses INTEGER,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_activo_estado CHECK (estado IN ('ACTIVO','BAJA')),
  CONSTRAINT fk_activo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_activo_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_activo ON inversiones.activo (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE inversiones.inversion (          -- inversión por obra y periodo
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID,
  periodo TEXT NOT NULL,                       -- YYYY-MM
  concepto TEXT NOT NULL,
  monto NUMERIC(16,2) NOT NULL DEFAULT 0,
  moneda TEXT NOT NULL DEFAULT 'USD',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_inversion_moneda CHECK (moneda IN ('USD','MXN')),
  CONSTRAINT fk_inversion_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_inversion_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_inversion_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT
);
CREATE INDEX ix_inversion_mina_periodo ON inversiones.inversion (id_empresa, id_mina, periodo);

CREATE TABLE inversiones.consumo_acero (      -- evento append-only: acero consumido
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID,
  fecha DATE NOT NULL,
  tipo_acero TEXT NOT NULL,                    -- barra helicoidal, malla, split set, etc.
  cantidad NUMERIC(14,3) NOT NULL,
  unidad TEXT NOT NULL DEFAULT 'kg',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_acero_cantidad CHECK (cantidad >= 0),
  CONSTRAINT fk_acero_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_acero_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_acero_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT
);
CREATE INDEX ix_acero_mina_fecha ON inversiones.consumo_acero (id_empresa, id_mina, fecha);
