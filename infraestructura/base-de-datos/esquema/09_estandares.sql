-- ============================================================
-- 09 · Estándares — metas de tiempos/movimientos y productividad
-- ============================================================
-- Metas contra las que se mide el desempeño operativo real:
-- disponibilidad, utilización, demoras y productividad por actividad.

CREATE TABLE estandares.estandar_tiempo (     -- meta de tiempos por tipo de equipo
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_tipo_equipo UUID NOT NULL,
  disponibilidad_meta_pct NUMERIC(6,2),        -- % disponible objetivo
  utilizacion_meta_pct NUMERIC(6,2),           -- % utilización objetivo
  demora_max_pct NUMERIC(6,2),                 -- % máximo de demora tolerado
  vigente_desde DATE NOT NULL,
  estado TEXT NOT NULL DEFAULT 'VIGENTE',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_estandar_tiempo_estado CHECK (estado IN ('VIGENTE','HISTORICO')),
  CONSTRAINT fk_estandar_tiempo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_estandar_tiempo_tipo_equipo FOREIGN KEY (id_tipo_equipo) REFERENCES catalogos.tipo_equipo(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_estandar_tiempo ON estandares.estandar_tiempo (id_empresa, id_tipo_equipo, vigente_desde) WHERE eliminado_en IS NULL;

CREATE TABLE estandares.estandar_productividad (  -- meta de productividad por actividad
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_actividad UUID NOT NULL,                  -- barrenación, rezagado, acarreo, anclaje, zarpeo, relleno
  id_tipo_equipo UUID,
  unidad TEXT NOT NULL,                         -- TON_HORA / METRO_HORA / TON_GUARDIA
  valor_meta NUMERIC(14,4) NOT NULL,
  vigente_desde DATE NOT NULL,
  estado TEXT NOT NULL DEFAULT 'VIGENTE',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_estandar_prod_unidad CHECK (unidad IN ('TON_HORA','METRO_HORA','TON_GUARDIA','METRO_GUARDIA')),
  CONSTRAINT ck_estandar_prod_estado CHECK (estado IN ('VIGENTE','HISTORICO')),
  CONSTRAINT fk_estandar_prod_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_estandar_prod_actividad FOREIGN KEY (id_actividad) REFERENCES catalogos.actividad(id) ON DELETE RESTRICT,
  CONSTRAINT fk_estandar_prod_tipo_equipo FOREIGN KEY (id_tipo_equipo) REFERENCES catalogos.tipo_equipo(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_estandar_prod ON estandares.estandar_productividad (id_empresa, id_actividad, vigente_desde) WHERE eliminado_en IS NULL;
