-- ============================================================
-- 01 · Gobierno y Acceso — tenant, usuarios, roles y superadmin
-- ============================================================
-- Modelo de administración (estándar B2B enterprise):
--   · Las EMPRESAS (tenants) las crea SOLO la plataforma (superadmin).
--     Una empresa NO puede crear otras empresas.
--   · Cada empresa tiene ADMIN(es) que dan de alta/baja a SUS usuarios
--     y les asignan roles (administración delegada). Sin auto-registro:
--     si alguien sale de la empresa, el admin lo da de baja y pierde acceso.
--   · El rol puede tener ALCANCE por mina (usuario_rol.id_mina) o aplicar
--     a toda la empresa (NULL).
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
COMMENT ON TABLE gobierno.empresa IS 'Tenant: empresa/grupo minero dueño de los datos. La crea SOLO la plataforma (superadmin).';

CREATE TABLE gobierno.usuario (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa                  UUID NOT NULL,
  id_empleado                 UUID,                -- vínculo opcional al personal (FK compuesta en 12)
  usuario                     TEXT NOT NULL,
  nombre                      TEXT NOT NULL,
  correo                      TEXT,
  contrasena_hash             TEXT,                -- hash (argon2/bcrypt); lo gestiona el backend
  ultimo_acceso_en            TIMESTAMPTZ,
  estado                      TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en                   TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por_usuario_id       UUID,
  actualizado_en              TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_por_usuario_id  UUID,
  eliminado_en                TIMESTAMPTZ,
  eliminado_por_usuario_id    UUID,
  CONSTRAINT ck_usuario_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT uq_usuario_tenant UNIQUE (id_empresa, id),
  CONSTRAINT fk_usuario_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_usuario ON gobierno.usuario (id_empresa, usuario) WHERE eliminado_en IS NULL;
CREATE INDEX ix_usuario_empleado ON gobierno.usuario (id_empleado) WHERE id_empleado IS NOT NULL;
COMMENT ON TABLE gobierno.usuario IS 'Cuenta de acceso por empresa. La dan de alta los admins de la empresa (sin auto-registro). Actor de la auditoria inline.';

-- ---------- RBAC: roles por empresa + asignación con alcance ----------
CREATE TABLE gobierno.rol (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa                  UUID NOT NULL,
  codigo                      TEXT NOT NULL,        -- ADMIN_EMPRESA, JEFE_TURNO, CAPITAN_MINA, OPERADOR, PLANEACION, LECTURA...
  descripcion                 TEXT NOT NULL,
  es_sistema                  BOOLEAN NOT NULL DEFAULT false,  -- roles semilla protegidos (no se editan/borran desde la app)
  estado                      TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en                   TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por_usuario_id       UUID,
  actualizado_en              TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_por_usuario_id  UUID,
  eliminado_en                TIMESTAMPTZ,
  eliminado_por_usuario_id    UUID,
  CONSTRAINT ck_rol_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT uq_rol_tenant UNIQUE (id_empresa, id),
  CONSTRAINT fk_rol_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_rol_codigo ON gobierno.rol (id_empresa, codigo) WHERE eliminado_en IS NULL;
COMMENT ON TABLE gobierno.rol IS 'Roles por empresa. Semilla de roles de sistema al crear el tenant; la empresa puede agregar roles propios.';

CREATE TABLE gobierno.usuario_rol (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa                  UUID NOT NULL,
  id_usuario                  UUID NOT NULL,
  id_rol                      UUID NOT NULL,
  id_mina                     UUID,                -- alcance: una mina específica o NULL = toda la empresa (FK compuesta en 12)
  creado_en                   TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por_usuario_id       UUID,
  actualizado_en              TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_por_usuario_id  UUID,
  eliminado_en                TIMESTAMPTZ,          -- la REVOCACIÓN es baja lógica: queda quién la hizo y cuándo
  eliminado_por_usuario_id    UUID,
  CONSTRAINT fk_usuario_rol_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_usuario_rol_usuario FOREIGN KEY (id_empresa, id_usuario) REFERENCES gobierno.usuario (id_empresa, id) ON DELETE RESTRICT,
  CONSTRAINT fk_usuario_rol_rol     FOREIGN KEY (id_empresa, id_rol)     REFERENCES gobierno.rol     (id_empresa, id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_usuario_rol ON gobierno.usuario_rol
  (id_empresa, id_usuario, id_rol, COALESCE(id_mina, '00000000-0000-0000-0000-000000000000'))
  WHERE eliminado_en IS NULL;
CREATE INDEX ix_usuario_rol_usuario ON gobierno.usuario_rol (id_usuario) WHERE eliminado_en IS NULL;
CREATE INDEX ix_usuario_rol_rol     ON gobierno.usuario_rol (id_rol) WHERE eliminado_en IS NULL;
COMMENT ON TABLE gobierno.usuario_rol IS 'Asignación usuario-rol con alcance opcional por mina. FKs compuestas: imposible asignar roles de otra empresa.';

-- ---------- Superadmin de plataforma (FUERA de los tenants) ----------
CREATE TABLE gobierno.superadmin (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  usuario                     TEXT NOT NULL,
  nombre                      TEXT NOT NULL,
  correo                      TEXT,
  contrasena_hash             TEXT,
  estado                      TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en                   TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por_usuario_id       UUID,
  actualizado_en              TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_por_usuario_id  UUID,
  eliminado_en                TIMESTAMPTZ,
  eliminado_por_usuario_id    UUID,
  CONSTRAINT ck_superadmin_estado CHECK (estado IN ('ACTIVO','INACTIVO'))
);
CREATE UNIQUE INDEX uq_superadmin ON gobierno.superadmin (usuario) WHERE eliminado_en IS NULL;
COMMENT ON TABLE gobierno.superadmin IS 'Operadores de la PLATAFORMA (nosotros): crean empresas y su primer admin. Invisible e inaccesible para el rol aplicacion.';
