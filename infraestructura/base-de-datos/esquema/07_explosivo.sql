-- ============================================================
-- 07 · Explosivo — catálogo + consumo (control de explosivo)
-- ============================================================
-- Informes de explosivo en desarrollo y tumbe: agentes, iniciadores,
-- cordón detonante, conectores.

CREATE TABLE catalogos.tipo_explosivo (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL,
  descripcion TEXT NOT NULL,
  familia TEXT NOT NULL DEFAULT 'AGENTE',     -- AGENTE / INICIADOR / CORDON / CONECTOR
  unidad_medida TEXT NOT NULL DEFAULT 'kg',
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_explosivo_familia CHECK (familia IN ('AGENTE','INICIADOR','CORDON','CONECTOR')),
  CONSTRAINT ck_tipo_explosivo_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_explosivo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_explosivo ON catalogos.tipo_explosivo (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE produccion.consumo_explosivo (   -- evento append-only
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID NOT NULL,
  id_tipo_explosivo UUID NOT NULL,
  id_parte_barrenacion UUID,                  -- opcional: liga al parte de barrenación
  fecha DATE NOT NULL,
  turno TEXT NOT NULL,
  destino TEXT NOT NULL DEFAULT 'DESARROLLO', -- DESARROLLO / TUMBE
  cantidad NUMERIC(14,3) NOT NULL,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_consumo_exp_turno CHECK (turno IN ('M','V','N')),
  CONSTRAINT ck_consumo_exp_destino CHECK (destino IN ('DESARROLLO','TUMBE')),
  CONSTRAINT ck_consumo_exp_cantidad CHECK (cantidad >= 0),
  CONSTRAINT fk_consumo_exp_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_consumo_exp_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_consumo_exp_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_consumo_exp_tipo FOREIGN KEY (id_tipo_explosivo) REFERENCES catalogos.tipo_explosivo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_consumo_exp_parte FOREIGN KEY (id_parte_barrenacion) REFERENCES produccion.parte_barrenacion(id) ON DELETE RESTRICT
);
CREATE INDEX ix_consumo_exp_mina_fecha ON produccion.consumo_explosivo (id_empresa, id_mina, fecha);
