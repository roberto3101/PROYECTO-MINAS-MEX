-- ============================================================
-- 40 · Escenario E2E — un ecosistema de usuarios reales
-- ============================================================
-- Simula el ciclo de vida completo del producto con los ROLES DE BD reales:
--   plataforma (nuestro panel superadmin)  → aprovisiona una empresa nueva
--   aplicacion (la app de los tenants)     → el admin configura y opera
-- Pasos: alta de empresa con branding → roles+permisos → admin → catálogos →
-- usuario operador con alcance por mina → captura de producción → aislamiento →
-- permisos efectivos → revocación y baja. Re-ejecutable (E0 limpia el tenant demo).

\set c_emp   'cccc1111-0000-0000-0000-00000000c001'
\set c_adm   'cccc1111-0000-0000-0000-00000000c002'
\set c_opu   'cccc1111-0000-0000-0000-00000000c003'
\set c_mina  'cccc1111-0000-0000-0000-00000000c010'
\set c_obra  'cccc1111-0000-0000-0000-00000000c011'
\set c_teq   'cccc1111-0000-0000-0000-00000000c012'
\set c_mod   'cccc1111-0000-0000-0000-00000000c013'
\set c_tmin  'cccc1111-0000-0000-0000-00000000c014'
\set c_tobra 'cccc1111-0000-0000-0000-00000000c015'
\set c_eq    'cccc1111-0000-0000-0000-00000000c016'
\set c_emple 'cccc1111-0000-0000-0000-00000000c017'
\set c_parte 'cccc1111-0000-0000-0000-00000000c020'
\set c_ur    'cccc1111-0000-0000-0000-00000000c030'

-- ===== E0 (mantenimiento): limpiar el tenant demo para re-ejecución =====
DELETE FROM produccion.acarreo_viaje      WHERE id_empresa = :'c_emp';
DELETE FROM produccion.parte_acarreo      WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.equipo              WHERE id_empresa = :'c_emp';
DELETE FROM gobierno.usuario_rol          WHERE id_empresa = :'c_emp';
DELETE FROM gobierno.usuario              WHERE id_empresa = :'c_emp';
DELETE FROM gobierno.rol_permiso          WHERE id_empresa = :'c_emp';
DELETE FROM gobierno.rol                  WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.obra                WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.empleado            WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.mina                WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.tipo_equipo         WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.modulo_trabajo      WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.tipo_mineral        WHERE id_empresa = :'c_emp';
DELETE FROM catalogos.tipo_obra           WHERE id_empresa = :'c_emp';
DELETE FROM gobierno.empresa              WHERE id = :'c_emp';

-- ===== E1 (plataforma): alta de la empresa con branding + roles + matriz =====
SET ROLE plataforma;
INSERT INTO gobierno.empresa (id, codigo, razon_social, identificacion_fiscal, correo_contacto,
                              logo_url, color_primario, zona_horaria, moneda_defecto)
VALUES (:'c_emp', 'MIN3', 'Minera Tres SA de CV', 'MTR240101XX1', 'contacto@mineratres.mx',
        'https://cdn.mineratres.mx/logo.png', '#0E7A4B', 'America/Chihuahua', 'MXN');
INSERT INTO gobierno.usuario (id, id_empresa, usuario, nombre, correo)
VALUES (:'c_adm', :'c_emp', 'admin.tres', 'ADMIN MINERA TRES', 'admin@mineratres.mx');
INSERT INTO gobierno.rol (id_empresa, codigo, descripcion, es_sistema)
SELECT :'c_emp', r.codigo, r.descripcion, true
FROM (VALUES
  ('ADMIN_EMPRESA','Administra usuarios, roles y catalogos de su empresa'),
  ('JEFE_TURNO','Supervisa la operacion del turno y valida partes'),
  ('CAPITAN_MINA','Captura y valida partes de su mina'),
  ('OPERADOR','Captura sus propios partes de operacion'),
  ('PLANEACION','Gestiona plan de bloques, metas y reportes'),
  ('LECTURA','Solo consulta tableros y reportes')
) AS r(codigo, descripcion);
INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)      -- ADMIN: todo el catalogo
SELECT :'c_emp', r.id, p.id FROM gobierno.rol r CROSS JOIN gobierno.permiso p
WHERE r.id_empresa = :'c_emp' AND r.codigo='ADMIN_EMPRESA';
INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)      -- OPERADOR: capturar y ver
SELECT :'c_emp', r.id, p.id FROM gobierno.rol r JOIN gobierno.permiso p ON p.codigo IN ('produccion.ver','produccion.capturar')
WHERE r.id_empresa = :'c_emp' AND r.codigo='OPERADOR';
INSERT INTO gobierno.usuario_rol (id_empresa, id_usuario, id_rol)      -- el primer admin
SELECT :'c_emp', :'c_adm', r.id FROM gobierno.rol r WHERE r.id_empresa = :'c_emp' AND r.codigo='ADMIN_EMPRESA';
RESET ROLE;
DO $$ BEGIN RAISE NOTICE 'OK  E1: plataforma aprovisiono MIN3 (branding+admin+6 roles+matriz) SIN tocar datos operativos'; END $$;

-- ===== E2 (admin de MIN3 en la app): configura catálogos y su equipo de trabajo =====
SELECT set_config('app.empresa_actual', :'c_emp', false);
SET ROLE aplicacion;
INSERT INTO catalogos.tipo_obra      (id, id_empresa, codigo, descripcion) VALUES (:'c_tobra', :'c_emp','REB','Rebaje');
INSERT INTO catalogos.tipo_equipo    (id, id_empresa, codigo, descripcion) VALUES (:'c_teq',   :'c_emp','CAMION','Camion');
INSERT INTO catalogos.modulo_trabajo (id, id_empresa, codigo, descripcion) VALUES (:'c_mod',   :'c_emp','ACARREO','Acarreo');
INSERT INTO catalogos.tipo_mineral   (id, id_empresa, codigo, descripcion, es_mineral) VALUES (:'c_tmin', :'c_emp','MINERAL','Mineral', true);
INSERT INTO catalogos.mina (id, id_empresa, nombre, area) VALUES (:'c_mina', :'c_emp','Mina Tres Norte','Norte');
INSERT INTO catalogos.obra (id, id_empresa, id_mina, id_tipo_obra, codigo, nombre)
VALUES (:'c_obra', :'c_emp', :'c_mina', :'c_tobra', 'REB-100','Rebaje 100');
INSERT INTO catalogos.equipo (id, id_empresa, id_mina, id_tipo_equipo, id_modulo_trabajo, codigo, fabricante)
VALUES (:'c_eq', :'c_emp', :'c_mina', :'c_teq', :'c_mod', 'C3-EQ01','Sandvik');
INSERT INTO catalogos.empleado (id, id_empresa, id_mina, numero_nomina, nombre_completo)
VALUES (:'c_emple', :'c_emp', :'c_mina', 'C3-001','OPERADOR DEMO TRES');
INSERT INTO gobierno.usuario (id, id_empresa, id_empleado, usuario, nombre)
VALUES (:'c_opu', :'c_emp', :'c_emple', 'op.tres', 'OPERADOR DEMO TRES');
INSERT INTO gobierno.usuario_rol (id, id_empresa, id_usuario, id_rol, id_mina)   -- rol con ALCANCE a su mina
SELECT :'c_ur', :'c_emp', :'c_opu', r.id, :'c_mina'
FROM gobierno.rol r WHERE r.id_empresa = :'c_emp' AND r.codigo='OPERADOR';
DO $$ BEGIN RAISE NOTICE 'OK  E2: admin configuro catalogos minimos y dio de alta a op.tres con rol OPERADOR alcance Mina Tres Norte'; END $$;

-- ===== E3 (op.tres captura su turno): parte + viajes con observaciones =====
INSERT INTO produccion.parte_acarreo (id, id_empresa, id_mina, id_obra, id_equipo, id_operador,
                                      fecha, turno, horometro_inicial, horometro_final, observaciones,
                                      creado_por_usuario_id)
VALUES (:'c_parte', :'c_emp', :'c_mina', :'c_obra', :'c_eq', :'c_emple',
        '2026-06-09','M', 100.0, 107.5, 'Primer turno de la mina demo; sin novedades.', :'c_opu');
INSERT INTO produccion.acarreo_viaje (id_empresa, id_parte_acarreo, desde, hasta, id_tipo_mineral, toneladas, creado_por_usuario_id)
VALUES (:'c_emp', :'c_parte', 'Tope 100','Patio', :'c_tmin', 35.5, :'c_opu'),
       (:'c_emp', :'c_parte', 'Tope 100','Patio', :'c_tmin', 36.2, :'c_opu');
DO $$ DECLARE v numeric; BEGIN
  SELECT sum(toneladas) INTO v FROM produccion.acarreo_viaje;
  IF v <> 71.7 THEN RAISE EXCEPTION 'E3 FALLO: toneladas esperadas 71.7, hay %', v; END IF;
  RAISE NOTICE 'OK  E3: op.tres capturo parte con observaciones y 2 viajes (71.7 t)';
END $$;

-- ===== E4: aislamiento — la empresa A no ve NADA de MIN3 =====
RESET ROLE;
SELECT set_config('app.empresa_actual', (SELECT id::text FROM gobierno.empresa WHERE codigo='MIN'), false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM catalogos.mina WHERE nombre='Mina Tres Norte';
  IF v <> 0 THEN RAISE EXCEPTION 'E4 FALLO: el tenant A ve la mina de MIN3'; END IF;
  RAISE NOTICE 'OK  E4: aislamiento intacto — MIN ve 0 datos de MIN3';
END $$;
RESET ROLE;

-- ===== E5: permisos efectivos del operador (lo que iria a su JWT) =====
SELECT set_config('app.empresa_actual', :'c_emp', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; v_alc int; BEGIN
  SELECT count(*), count(*) FILTER (WHERE alcance_mina IS NOT NULL)
    INTO v, v_alc FROM gobierno.v_permisos_usuario WHERE usuario='op.tres';
  IF v <> 2 THEN RAISE EXCEPTION 'E5 FALLO: op.tres debe tener 2 permisos, tiene %', v; END IF;
  IF v_alc <> 2 THEN RAISE EXCEPTION 'E5 FALLO: el alcance por mina no llego a los permisos (%/2)', v_alc; END IF;
  RAISE NOTICE 'OK  E5: op.tres con exactamente 2 permisos (produccion.ver/capturar) y alcance por mina';
END $$;

-- ===== E6: lo corren — el admin revoca el rol y da de baja al usuario =====
UPDATE gobierno.usuario_rol SET eliminado_en = now(), eliminado_por_usuario_id = :'c_adm' WHERE id = :'c_ur';
UPDATE gobierno.usuario SET estado='INACTIVO', actualizado_por_usuario_id = :'c_adm' WHERE id = :'c_opu';
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM gobierno.v_permisos_usuario WHERE usuario='op.tres';
  IF v <> 0 THEN RAISE EXCEPTION 'E6 FALLO: op.tres conserva % permisos tras la revocacion', v; END IF;
  RAISE NOTICE 'OK  E6: revocacion trazable — op.tres quedo sin permisos al instante (y la captura historica persiste)';
END $$;
RESET ROLE;
SELECT set_config('app.empresa_actual', '', false);

DO $$ BEGIN RAISE NOTICE '=============================================='; RAISE NOTICE 'ESCENARIO DE USUARIOS REALES: COMPLETO'; RAISE NOTICE '=============================================='; END $$;
