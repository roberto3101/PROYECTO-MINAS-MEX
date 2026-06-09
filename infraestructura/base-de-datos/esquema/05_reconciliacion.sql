-- ============================================================
-- 05 · Reconciliación — trazabilidad por segmentos 10x2 m
-- ============================================================
-- Cada segmento (10 m x 2 m) tiene identificador único que se persigue
-- por toda la cadena: Reserva -> Tumbe -> Estimación -> Planta.

CREATE TABLE reconciliacion.segmento (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_rebaje UUID,
  codigo TEXT NOT NULL,                       -- identificador único del segmento
  largo_m NUMERIC(6,2) NOT NULL DEFAULT 10.0,
  ancho_m NUMERIC(6,2) NOT NULL DEFAULT 2.0,
  descripcion TEXT,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_segmento_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_segmento_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_segmento_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_segmento_rebaje FOREIGN KEY (id_rebaje) REFERENCES planeacion.rebaje(id) ON DELETE RESTRICT
);
CREATE INDEX ix_segmento_mina ON reconciliacion.segmento (id_empresa, id_mina);
CREATE UNIQUE INDEX uq_segmento ON reconciliacion.segmento (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE reconciliacion.medicion (        -- evento append-only: una medición por fuente
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_segmento UUID NOT NULL,
  fuente TEXT NOT NULL,                       -- RESERVA / TUMBE / ESTIMACION / PLANTA
  fecha DATE NOT NULL,
  toneladas NUMERIC(14,3),
  ley_au NUMERIC(10,4),
  ley_ag NUMERIC(10,4),
  ley_pb NUMERIC(10,4),
  ley_zn NUMERIC(10,4),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_medicion_fuente CHECK (fuente IN ('RESERVA','TUMBE','ESTIMACION','PLANTA')),
  CONSTRAINT uq_medicion UNIQUE (id_empresa, id_segmento, fuente),
  CONSTRAINT fk_medicion_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_medicion_segmento FOREIGN KEY (id_segmento) REFERENCES reconciliacion.segmento(id) ON DELETE RESTRICT
);
CREATE INDEX ix_medicion_segmento ON reconciliacion.medicion (id_segmento);
