-- ============================================================
-- 02 · Catálogos — datos maestros de la operación
-- ============================================================
-- Patrón lookup: codigo (clave natural por empresa) + descripcion + estado.
-- Reemplazan los campos de texto libre / listas fijas del ER original.
-- Unicidad por índice PARCIAL (solo activos): permite re-alta tras baja lógica.

-- ---------- Lookups (tipificaciones) ----------
CREATE TABLE catalogos.tipo_obra (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL,           -- RPA, LAT, ACG, CRO, REB, STOCK, STILL
  descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_obra_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_obra_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_obra ON catalogos.tipo_obra (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.sistema_minado (   -- campo "Tumbe" del ER: BL, CR, TSC
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_sistema_minado_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_sistema_minado_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_sistema_minado ON catalogos.sistema_minado (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.tipo_equipo (   -- antes "Tipo" texto libre: CAMION, SCOOP, JUMBO_LINEAL...
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_equipo_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_equipo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_equipo ON catalogos.tipo_equipo (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.modulo_trabajo (   -- ACARREO, REZAGADO, BARRENACION, ANCLAJE...
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_modulo_trabajo_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_modulo_trabajo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_modulo_trabajo ON catalogos.modulo_trabajo (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.tipo_mineral (   -- MINERAL, NO_MINERAL, TEPETATE
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  es_mineral BOOLEAN NOT NULL DEFAULT true,   -- regla: cuenta como produccion vs esteril
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_mineral_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_mineral_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_mineral ON catalogos.tipo_mineral (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.departamento (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_departamento_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_departamento_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_departamento ON catalogos.departamento (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.puesto (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_puesto_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_puesto_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_puesto ON catalogos.puesto (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.actividad (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_actividad_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_actividad_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_actividad ON catalogos.actividad (id_empresa, codigo) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.tipo_demora (   -- para estandares de tiempos/movimientos (demoras)
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  es_programada BOOLEAN NOT NULL DEFAULT false,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_demora_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_demora_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_demora ON catalogos.tipo_demora (id_empresa, codigo) WHERE eliminado_en IS NULL;

-- ---------- Maestros (Raíz de Agregado / Entidad) ----------
CREATE TABLE catalogos.mina (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  nombre TEXT NOT NULL,                 -- antes PK natural NombreMina
  area TEXT NOT NULL,                   -- antes se copiaba a empleados/produccion
  niveles TEXT,
  densidad_promedio NUMERIC(6,3),
  seccion_obra_largo NUMERIC(7,2),
  seccion_obra_ancho NUMERIC(7,2),
  seccion_veta_largo NUMERIC(7,2),
  seccion_veta_ancho NUMERIC(7,2),
  estado TEXT NOT NULL DEFAULT 'ACTIVA',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_mina_estado CHECK (estado IN ('ACTIVA','INACTIVA')),
  CONSTRAINT fk_mina_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_mina_nombre ON catalogos.mina (id_empresa, nombre) WHERE eliminado_en IS NULL;
COMMENT ON TABLE catalogos.mina IS 'Unidad minera. Raiz del agregado de catalogos. NombreMina (PK natural del ER) -> id UUID + nombre unico entre activas.';

CREATE TABLE catalogos.equipo (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_tipo_equipo UUID NOT NULL,
  id_modulo_trabajo UUID NOT NULL,
  codigo TEXT,                          -- codigo interno legible (antes Id NUMBER)
  descripcion TEXT,
  modelo_perforadora TEXT,
  capacidad_longitud NUMERIC(8,2),
  fabricante TEXT,
  modelo TEXT,
  numero_serie TEXT,
  anio_fabricacion SMALLINT,
  fecha_ingreso_mina DATE,
  estado TEXT NOT NULL DEFAULT 'OPERATIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_equipo_estado CHECK (estado IN ('OPERATIVO','MANTENIMIENTO','BAJA')),
  CONSTRAINT ck_equipo_anio CHECK (anio_fabricacion IS NULL OR anio_fabricacion BETWEEN 1900 AND 2100),
  CONSTRAINT fk_equipo_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_equipo_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_equipo_tipo FOREIGN KEY (id_tipo_equipo) REFERENCES catalogos.tipo_equipo(id) ON DELETE RESTRICT,
  CONSTRAINT fk_equipo_modulo FOREIGN KEY (id_modulo_trabajo) REFERENCES catalogos.modulo_trabajo(id) ON DELETE RESTRICT
);
CREATE INDEX ix_equipo_mina ON catalogos.equipo (id_empresa, id_mina);
CREATE UNIQUE INDEX uq_equipo_codigo ON catalogos.equipo (id_empresa, codigo) WHERE eliminado_en IS NULL AND codigo IS NOT NULL;

CREATE TABLE catalogos.empleado (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  numero_nomina TEXT NOT NULL,          -- antes PK natural Nomina
  nombre_completo TEXT NOT NULL,
  id_departamento UUID,
  id_puesto UUID,
  centro_costo TEXT,
  gerente_a_cargo TEXT,
  id_actividad UUID,
  grupo TEXT,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_empleado_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_empleado_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_empleado_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_empleado_departamento FOREIGN KEY (id_departamento) REFERENCES catalogos.departamento(id) ON DELETE RESTRICT,
  CONSTRAINT fk_empleado_puesto FOREIGN KEY (id_puesto) REFERENCES catalogos.puesto(id) ON DELETE RESTRICT,
  CONSTRAINT fk_empleado_actividad FOREIGN KEY (id_actividad) REFERENCES catalogos.actividad(id) ON DELETE RESTRICT
);
CREATE INDEX ix_empleado_mina ON catalogos.empleado (id_empresa, id_mina);
CREATE UNIQUE INDEX uq_empleado_nomina ON catalogos.empleado (id_empresa, numero_nomina) WHERE eliminado_en IS NULL;

CREATE TABLE catalogos.mineral (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_tipo_mineral UUID NOT NULL,
  nombre TEXT NOT NULL,                 -- Plata, Oro, Plomo, Zinc, Tepetate
  unidad_medida TEXT NOT NULL,          -- g/t, ton, %
  ley NUMERIC(10,4),
  observaciones TEXT,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_mineral_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_mineral_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_mineral_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_mineral_tipo FOREIGN KEY (id_tipo_mineral) REFERENCES catalogos.tipo_mineral(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_mineral ON catalogos.mineral (id_empresa, id_mina, nombre) WHERE eliminado_en IS NULL;

-- ---------- Obra (entidad central del dominio: la labor minera) ----------
-- Antes era texto suelto en cada parte; ahora es entidad de primera clase.
CREATE TABLE catalogos.obra (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  id_mina UUID NOT NULL,
  id_tipo_obra UUID,
  codigo TEXT NOT NULL,                  -- p.ej. REB-2400W, RPA-2140
  nombre TEXT,
  ubicacion TEXT,                        -- nivel / zona dentro de la mina
  es_prioritaria BOOLEAN NOT NULL DEFAULT false,   -- seguimiento de obras prioritarias
  estado TEXT NOT NULL DEFAULT 'ACTIVA',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_obra_estado CHECK (estado IN ('ACTIVA','SUSPENDIDA','CONCLUIDA')),
  CONSTRAINT fk_obra_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT,
  CONSTRAINT fk_obra_mina FOREIGN KEY (id_mina) REFERENCES catalogos.mina(id) ON DELETE RESTRICT,
  CONSTRAINT fk_obra_tipo_obra FOREIGN KEY (id_tipo_obra) REFERENCES catalogos.tipo_obra(id) ON DELETE RESTRICT
);
CREATE INDEX ix_obra_mina ON catalogos.obra (id_empresa, id_mina);
CREATE INDEX ix_obra_prioritaria ON catalogos.obra (id_empresa, es_prioritaria) WHERE es_prioritaria;
CREATE UNIQUE INDEX uq_obra ON catalogos.obra (id_empresa, id_mina, codigo) WHERE eliminado_en IS NULL;
COMMENT ON TABLE catalogos.obra IS 'Labor minera (RPA, LAT, REB...). Termino central del dominio. Referenciada por los partes de produccion.';

-- ---------- Tipo de barreno (secciones de la voladura de frente) ----------
-- Modela el campo "Ejecutado" (lista fija múltiple) del ER: las secciones de
-- barreno de una frente de desarrollo (cuele, destroza, contorno, zapatera...).
CREATE TABLE catalogos.tipo_barreno (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  id_empresa UUID NOT NULL,
  codigo TEXT NOT NULL,                  -- CUELE, CONTRACUELE, DESTROZA, CONTORNO, ZAPATERA, CORONA, AUXILIAR
  descripcion TEXT NOT NULL,
  estado TEXT NOT NULL DEFAULT 'ACTIVO',
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), creado_por_usuario_id UUID,
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(), actualizado_por_usuario_id UUID,
  eliminado_en TIMESTAMPTZ, eliminado_por_usuario_id UUID,
  CONSTRAINT ck_tipo_barreno_estado CHECK (estado IN ('ACTIVO','INACTIVO')),
  CONSTRAINT fk_tipo_barreno_empresa FOREIGN KEY (id_empresa) REFERENCES gobierno.empresa(id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX uq_tipo_barreno ON catalogos.tipo_barreno (id_empresa, codigo) WHERE eliminado_en IS NULL;
