-- ============================================================
-- 01 · Gobierno central — empresa (tenant) y usuario (auditoría)
-- ============================================================
-- Mínimo necesario para multi-tenant y FKs. En el sistema real
-- esta es una capacidad completa (roles, permisos, periodos...).
-- Unicidad por índice PARCIAL (solo activos) para permitir re-alta tras baja lógica.

CREATE TABLE gobierno.empresa (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  codigo                      TEXT NOT NULL,
  razon_social                TEXT NOT NULL,
  estado                      TEXT NOT NULL DEFAULT 'ACTIVA',
  creado_en                   TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por_usuario_id       UUID,
  actualizado_en              TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_por_usuario_id  UUID,
  eliminado_en                TIMESTAMPTZ,
  eliminado_por_usuario_id    UUID,
  CONSTRAINT ck_empresa_estado CHECK (estado IN ('ACTIVA','INACTIVA'))
);
CREATE UNIQUE INDEX uq_empresa_codigo ON gobierno.empresa (codigo) WHERE eliminado_en IS NULL;
COMMENT ON TABLE gobierno.empresa IS 'Tenant: empresa/grupo minero dueño de los datos. Raiz de la multi-tenencia.';

CREATE TABLE gobierno.usuario (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa                  UUID NOT NULL,
  usuario                     TEXT NOT NULL,
  nombre                      TEXT NOT NULL,
  correo                      TEXT,
  estado                      TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en                   TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por_usuario_id       UUID,
  actualizado_en              TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_por_usuario_id  UUID,
  eliminado_en                TIMESTAMPTZ,
  eliminado_por_usuario_id    UUID,
  CONSTRAINT ck_usuario_estado CHECK (estado IN ('ACTIVO','INACTIVO'))
);
ALTER TABLE gobierno.usuario
  ADD CONSTRAINT fk_usuario_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT;
CREATE UNIQUE INDEX uq_usuario ON gobierno.usuario (id_empresa, usuario) WHERE eliminado_en IS NULL;
COMMENT ON TABLE gobierno.usuario IS 'Usuario operativo/administrativo por empresa. Actor de la auditoria inline.';
