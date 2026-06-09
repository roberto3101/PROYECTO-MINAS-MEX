-- ============================================================
-- 60 · Seguridad multi-tenant — RLS, rol de aplicación y privilegios
-- ============================================================
-- El aislamiento vive EN LA BASE; el backend solo fija el tenant:
--
--   BEGIN;
--   SELECT set_config('app.empresa_actual', $1, true);  -- uuid del JWT (true = local a la tx)
--   SET LOCAL ROLE aplicacion;
--   ... consultas normales, sin WHERE id_empresa ...
--   COMMIT;
--
-- Capas:
--   (1) Rol `aplicacion` sin BYPASSRLS: todo pasa por las políticas.
--   (2) RLS fail-closed en TODA tabla con id_empresa: sin tenant fijado -> 0 filas.
--   (3) Privilegios que refuerzan la metodología: eventos sin UPDATE
--       (append-only) y NADIE de la app con DELETE (borrado lógico).
--   (4) Vistas con security_invoker: el RLS del que consulta atraviesa la vista.
--   (5) La materializada no soporta RLS -> se expone solo vía vista filtrada.
-- Se ejecuta DESPUÉS de 50_vistas. Corre como superusuario (bypassa RLS),
-- por eso semilla y pruebas de lógica no se ven afectadas.

-- ---------- (1) Rol de aplicación ----------
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'aplicacion') THEN
    CREATE ROLE aplicacion NOLOGIN;
    COMMENT ON ROLE aplicacion IS 'Rol del backend: sujeto a RLS, sin DELETE, sin UPDATE en eventos. El rol LOGIN real del backend debe ser miembro de este (CREATE ROLE backend LOGIN ... IN ROLE aplicacion).';
  END IF;
END $$;

GRANT USAGE ON SCHEMA gobierno, catalogos, produccion, planeacion, reconciliacion,
                      beneficio, estandares, costos, inversiones, reportes TO aplicacion;

-- ---------- (3) Privilegios por tipo de tabla ----------
-- Mutables (tienen actualizado_en): SELECT, INSERT, UPDATE (el "borrado" es UPDATE de eliminado_en).
-- Eventos (solo creado_*):          SELECT, INSERT — append-only también por privilegios.
DO $$
DECLARE t RECORD;
BEGIN
  FOR t IN
    SELECT tb.table_schema AS s, tb.table_name AS n,
           EXISTS (SELECT 1 FROM information_schema.columns c
                   WHERE c.table_schema = tb.table_schema
                     AND c.table_name  = tb.table_name
                     AND c.column_name = 'actualizado_en') AS es_mutable
    FROM information_schema.tables tb
    WHERE tb.table_type = 'BASE TABLE'
      AND tb.table_schema IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
  LOOP
    IF t.es_mutable THEN
      EXECUTE format('GRANT SELECT, INSERT, UPDATE ON %I.%I TO aplicacion', t.s, t.n);
    ELSE
      EXECUTE format('GRANT SELECT, INSERT ON %I.%I TO aplicacion', t.s, t.n);
    END IF;
  END LOOP;
END $$;

GRANT SELECT ON gobierno.empresa TO aplicacion;            -- los tenants los administra la plataforma, no la app
GRANT SELECT, INSERT, UPDATE ON gobierno.usuario TO aplicacion;
GRANT SELECT, INSERT, UPDATE ON gobierno.rol TO aplicacion;          -- el admin de empresa gestiona sus roles
GRANT SELECT, INSERT, UPDATE ON gobierno.usuario_rol TO aplicacion;  -- ...y las asignaciones (revocación = baja lógica)
-- gobierno.superadmin: SIN grants para aplicacion (invisible para los tenants)

-- ---------- Rol de PLATAFORMA (superadmin: nosotros) ----------
-- Administra tenants: crea empresas, su primer admin y los roles semilla.
-- Es la ÚNICA vía de la aplicación de plataforma; también está sujeta a RLS
-- pero con política propia (ve todos los tenants SOLO en tablas de gobierno).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'plataforma') THEN
    CREATE ROLE plataforma NOLOGIN;
    COMMENT ON ROLE plataforma IS 'Rol del panel superadmin de la plataforma: gestiona empresas, admins iniciales y roles. No accede a datos operativos.';
  END IF;
END $$;
GRANT USAGE ON SCHEMA gobierno TO plataforma;
GRANT SELECT, INSERT, UPDATE ON gobierno.empresa, gobierno.usuario, gobierno.rol,
                               gobierno.usuario_rol, gobierno.superadmin TO plataforma;

-- ---------- (2) RLS fail-closed en toda tabla con id_empresa ----------
DO $$
DECLARE t RECORD;
BEGIN
  FOR t IN
    SELECT c.table_schema AS s, c.table_name AS n
    FROM information_schema.columns c
    JOIN information_schema.tables tb
      ON tb.table_schema = c.table_schema AND tb.table_name = c.table_name AND tb.table_type = 'BASE TABLE'
    WHERE c.column_name = 'id_empresa'
      AND c.table_schema IN ('gobierno','catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
  LOOP
    EXECUTE format('ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY', t.s, t.n);
    EXECUTE format('ALTER TABLE %I.%I FORCE ROW LEVEL SECURITY', t.s, t.n);
    EXECUTE format('DROP POLICY IF EXISTS p_tenant ON %I.%I', t.s, t.n);
    EXECUTE format($p$CREATE POLICY p_tenant ON %I.%I
                     USING      (id_empresa = NULLIF(current_setting('app.empresa_actual', true), '')::uuid)
                     WITH CHECK (id_empresa = NULLIF(current_setting('app.empresa_actual', true), '')::uuid)$p$,
                   t.s, t.n);
  END LOOP;
END $$;

-- empresa (no tiene id_empresa: ES el tenant): cada empresa solo se ve a sí misma.
ALTER TABLE gobierno.empresa ENABLE ROW LEVEL SECURITY;
ALTER TABLE gobierno.empresa FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS p_tenant ON gobierno.empresa;
CREATE POLICY p_tenant ON gobierno.empresa
  USING (id = NULLIF(current_setting('app.empresa_actual', true), '')::uuid);

-- superadmin: tabla de plataforma, fuera de los tenants. Solo el rol plataforma.
ALTER TABLE gobierno.superadmin ENABLE ROW LEVEL SECURITY;
ALTER TABLE gobierno.superadmin FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS p_plataforma ON gobierno.superadmin;
CREATE POLICY p_plataforma ON gobierno.superadmin TO plataforma USING (true) WITH CHECK (true);

-- El panel de plataforma ve y administra TODOS los tenants, pero SOLO en gobierno
-- (las políticas son permisivas: se suman con OR a p_tenant para este rol).
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['empresa','usuario','rol','usuario_rol']
  LOOP
    EXECUTE format('DROP POLICY IF EXISTS p_plataforma ON gobierno.%I', t);
    EXECUTE format('CREATE POLICY p_plataforma ON gobierno.%I TO plataforma USING (true) WITH CHECK (true)', t);
  END LOOP;
END $$;

-- ---------- (4) Vistas: el RLS del consultante atraviesa la vista ----------
DO $$
DECLARE v RECORD;
BEGIN
  FOR v IN SELECT viewname FROM pg_views WHERE schemaname = 'reportes'
  LOOP
    EXECUTE format('ALTER VIEW reportes.%I SET (security_invoker = on)', v.viewname);
  END LOOP;
END $$;

-- ---------- (5) Materializada: solo accesible vía vista filtrada ----------
-- (las materializadas no soportan RLS; esta vista corre como su dueño y
--  filtra por el tenant de la sesión — la app consulta ESTA, no la mv_).
CREATE OR REPLACE VIEW reportes.v_plan_vs_real_mensual AS
SELECT *
FROM reportes.mv_plan_vs_real_mensual
WHERE id_empresa = NULLIF(current_setting('app.empresa_actual', true), '')::uuid;
COMMENT ON VIEW reportes.v_plan_vs_real_mensual IS
  'Acceso tenant-seguro al plan vs real mensual. La app consulta esta vista; la materializada queda sin permisos para aplicacion.';

GRANT SELECT ON ALL TABLES IN SCHEMA reportes TO aplicacion;
REVOKE ALL ON reportes.mv_plan_vs_real_mensual FROM aplicacion;
