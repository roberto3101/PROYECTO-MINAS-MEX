-- ============================================================
-- 10 · Seguridad — incidentes y observaciones de seguridad
-- ============================================================
-- Capacidad de datos puros (sin hardware): captura de incidentes y
-- casi-pérdidas en mina. Es lo que toda operación minera registra y
-- complementa las demoras de producción con la dimensión de seguridad.
-- Patrón cabecera/evento del resto del sistema:
--   tipo_incidente = catálogo por empresa (mutable)
--   incidente      = evento append-only (solo creado_*; sin UPDATE/DELETE)
-- Se ejecuta después de catálogos/producción (usa mina, obra, empleado)
-- y antes de 11_indices / 12_integridad / 60_seguridad_rls.

CREATE TABLE seguridad.tipo_incidente (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL,                 -- CASIPERDIDA, LESION, CAIDA_ROCA, DERRAME, ELECTRICO, INCENDIO, VEHICULO
  descripcion TEXT NOT NULL,
  requiere_paro BOOLEAN NOT NULL DEFAULT false,   -- si obliga a detener la operación
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_incidente_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_incidente_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_incidente ON seguridad.tipo_incidente (id_empresa, codigo) WHERE eliminado_en IS NULL;
COMMENT ON TABLE seguridad.tipo_incidente IS 'Catalogo por empresa de tipos de incidente de seguridad.';

CREATE TABLE seguridad.incidente (    -- evento append-only: un hecho de seguridad ya ocurrido
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_obra UUID,                        -- opcional: dónde ocurrió, si aplica a una obra
  id_tipo_incidente UUID NOT NULL,
  id_reportado_por UUID,              -- empleado que reporta (opcional)
  fecha DATE NOT NULL,
  turno TEXT NOT NULL,                 -- M / V / N
  severidad TEXT NOT NULL,            -- BAJA / MEDIA / ALTA / CRITICA
  descripcion TEXT NOT NULL,
  accion_inmediata TEXT,             -- qué se hizo en el momento
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  CONSTRAINT ck_incidente_turno CHECK (turno IN ('M','V','N')),
  CONSTRAINT ck_incidente_severidad CHECK (severidad IN ('BAJA','MEDIA','ALTA','CRITICA')),
  CONSTRAINT fk_incidente_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_incidente_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_incidente_obra FOREIGN KEY (id_obra) REFERENCES catalogos.obra(id) ON DELETE RESTRICT,
  CONSTRAINT fk_incidente_tipo FOREIGN KEY (id_tipo_incidente) REFERENCES seguridad.tipo_incidente(id) ON DELETE RESTRICT,
  CONSTRAINT fk_incidente_reportado_por FOREIGN KEY (id_reportado_por) REFERENCES catalogos.empleado(id) ON DELETE RESTRICT
);
CREATE INDEX ix_incidente_mina_fecha ON seguridad.incidente (id_empresa, id_mina, fecha);
CREATE INDEX ix_incidente_tipo ON seguridad.incidente (id_tipo_incidente);
CREATE INDEX ix_incidente_obra ON seguridad.incidente (id_obra) WHERE id_obra IS NOT NULL;
CREATE INDEX ix_incidente_severidad ON seguridad.incidente (id_empresa, severidad, fecha);
COMMENT ON TABLE seguridad.incidente IS 'Evento append-only: incidente o casi-perdida de seguridad capturado en mina.';
