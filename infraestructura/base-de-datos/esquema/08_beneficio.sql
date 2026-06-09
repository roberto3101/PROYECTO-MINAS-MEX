-- ============================================================
-- 08 · Beneficio — planta de molienda (balance metalúrgico)
-- ============================================================
-- Soporta la comparativa Producción vs Molienda y la variación de
-- ley Tumbe vs Planta: leyes cabeza/concentrado/cola y recuperación.

CREATE TABLE beneficio.lote_molienda (        -- lote procesado en planta
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  codigo_lote TEXT NOT NULL,
  fecha DATE NOT NULL,
  toneladas_molidas NUMERIC(14,3) NOT NULL DEFAULT 0,
  estado TEXT NOT NULL DEFAULT 'ABIERTO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_lote_estado CHECK (estado IN ('ABIERTO','CERRADO')),
  CONSTRAINT fk_lote_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_lote_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT
);
CREATE INDEX ix_lote_mina_fecha ON beneficio.lote_molienda (id_empresa, id_mina, fecha);
CREATE UNIQUE INDEX uq_lote_molienda ON beneficio.lote_molienda (id_empresa, id_mina, codigo_lote) WHERE eliminado_en IS NULL;

CREATE TABLE beneficio.ley_metalurgica (      -- evento: ley por punto del proceso
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_lote_molienda UUID NOT NULL,
  punto TEXT NOT NULL,                         -- CABEZA / CONCENTRADO / COLA
  ley_au NUMERIC(10,4),
  ley_ag NUMERIC(10,4),
  ley_pb NUMERIC(10,4),
  ley_zn NUMERIC(10,4),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_ley_punto CHECK (punto IN ('CABEZA','CONCENTRADO','COLA')),
  CONSTRAINT uq_ley_metalurgica UNIQUE (id_empresa, id_lote_molienda, punto),
  CONSTRAINT fk_ley_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_ley_lote FOREIGN KEY (id_lote_molienda) REFERENCES beneficio.lote_molienda(id) ON DELETE RESTRICT
);
CREATE INDEX ix_ley_lote ON beneficio.ley_metalurgica (id_lote_molienda);

CREATE TABLE beneficio.recuperacion (         -- evento: recuperación metalúrgica por metal
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_lote_molienda UUID NOT NULL,
  metal TEXT NOT NULL,                         -- AU / AG / PB / ZN
  recuperacion_pct NUMERIC(6,2),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_recuperacion_metal CHECK (metal IN ('AU','AG','PB','ZN')),
  CONSTRAINT ck_recuperacion_pct CHECK (recuperacion_pct IS NULL OR (recuperacion_pct >= 0 AND recuperacion_pct <= 100)),
  CONSTRAINT uq_recuperacion UNIQUE (id_empresa, id_lote_molienda, metal),
  CONSTRAINT fk_recuperacion_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_recuperacion_lote FOREIGN KEY (id_lote_molienda) REFERENCES beneficio.lote_molienda(id) ON DELETE RESTRICT
);
CREATE INDEX ix_recuperacion_lote ON beneficio.recuperacion (id_lote_molienda);
