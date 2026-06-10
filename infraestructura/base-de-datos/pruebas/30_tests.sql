-- ============================================================
-- 30 · Tests de integración y lógica
-- ============================================================
-- Ejecutar con: psql -v ON_ERROR_STOP=1 -f 30_tests.sql
-- Un RAISE EXCEPTION en cualquier bloque aborta (test fallido).

REFRESH MATERIALIZED VIEW reportes.mv_plan_vs_real_mensual;

-- ===== INTEGRIDAD: constraints rechazan datos inválidos =====
DO $$ BEGIN
  BEGIN
    INSERT INTO produccion.parte_acarreo (id_empresa, id_mina, id_obra, id_equipo, id_operador, fecha, turno, horometro_inicial, horometro_final)
    SELECT id_empresa, id_mina, id_obra, id_equipo, id_operador, fecha, 'X', horometro_inicial, horometro_final
    FROM produccion.parte_acarreo LIMIT 1;
    RAISE EXCEPTION 'T1 FALLO: turno invalido fue aceptado';
  EXCEPTION WHEN check_violation THEN RAISE NOTICE 'OK  T1: turno invalido (X) rechazado por CHECK';
  END;
END $$;

DO $$ BEGIN
  BEGIN
    INSERT INTO produccion.parte_acarreo (id_empresa, id_mina, id_obra, id_equipo, id_operador, fecha, turno, horometro_inicial, horometro_final)
    SELECT id_empresa, id_mina, id_obra, id_equipo, id_operador, fecha, 'M', 100, 50
    FROM produccion.parte_acarreo LIMIT 1;
    RAISE EXCEPTION 'T2 FALLO: horometro_final < inicial fue aceptado';
  EXCEPTION WHEN check_violation THEN RAISE NOTICE 'OK  T2: horometro_final < inicial rechazado por CHECK';
  END;
END $$;

DO $$ BEGIN
  BEGIN
    INSERT INTO produccion.parte_barrenacion (id_empresa, tipo_barrenacion, id_mina, id_obra, id_equipo, id_capitan_mina, id_operador, fecha, turno)
    SELECT id_empresa,'INVALIDO', id_mina, id_obra, id_equipo, id_capitan_mina, id_operador, fecha, turno
    FROM produccion.parte_barrenacion LIMIT 1;
    RAISE EXCEPTION 'T3 FALLO: tipo_barrenacion invalido fue aceptado';
  EXCEPTION WHEN check_violation THEN RAISE NOTICE 'OK  T3: tipo_barrenacion invalido rechazado por CHECK';
  END;
END $$;

DO $$ BEGIN
  BEGIN
    INSERT INTO catalogos.equipo (id_empresa, id_mina, id_tipo_equipo, id_modulo_trabajo, codigo)
    VALUES ((SELECT id FROM gobierno.empresa WHERE codigo='MIN'), '00000000-0000-0000-0000-000000000000',
            (SELECT te.id FROM catalogos.tipo_equipo te JOIN gobierno.empresa e ON e.id=te.id_empresa WHERE e.codigo='MIN' LIMIT 1),
            (SELECT mt.id FROM catalogos.modulo_trabajo mt JOIN gobierno.empresa e ON e.id=mt.id_empresa WHERE e.codigo='MIN' LIMIT 1), 'EQ-FK');
    RAISE EXCEPTION 'T4 FALLO: FK a mina inexistente fue aceptada';
  EXCEPTION WHEN foreign_key_violation THEN RAISE NOTICE 'OK  T4: FK a mina inexistente rechazada';
  END;
END $$;

DO $$ BEGIN
  BEGIN
    INSERT INTO catalogos.mina (id_empresa, nombre, area)
    VALUES ((SELECT id FROM gobierno.empresa WHERE codigo='MIN'), 'La Cienega', 'Cienega');
    RAISE EXCEPTION 'T5 FALLO: nombre de mina duplicado fue aceptado';
  EXCEPTION WHEN unique_violation THEN RAISE NOTICE 'OK  T5: nombre de mina duplicado rechazado (indice unico parcial)';
  END;
END $$;

-- ===== LÓGICA: las vistas devuelven los valores esperados =====
DO $$ DECLARE v_min numeric; v_tep numeric; BEGIN
  SELECT ton_mineral, ton_tepetate INTO v_min, v_tep
  FROM reportes.v_balance_cargas WHERE mina='La Cienega' AND periodo_mes='2026-06';
  IF v_min <> 760 THEN RAISE EXCEPTION 'T6 FALLO: ton_mineral esperado 760, obtenido %', v_min; END IF;
  IF v_tep <> 200 THEN RAISE EXCEPTION 'T6 FALLO: ton_tepetate esperado 200, obtenido %', v_tep; END IF;
  RAISE NOTICE 'OK  T6: balance de cargas mineral=% tepetate=%', v_min, v_tep;
END $$;

DO $$ DECLARE v_plan numeric; v_real numeric; v_apego numeric; BEGIN
  SELECT mv.ton_plan, mv.ton_real, mv.apego_pct INTO v_plan, v_real, v_apego
  FROM reportes.mv_plan_vs_real_mensual mv JOIN catalogos.mina m ON m.id = mv.id_mina
  WHERE m.nombre='La Cienega' AND mv.periodo='2026-06';
  IF v_plan <> 1000 THEN RAISE EXCEPTION 'T7 FALLO: ton_plan esperado 1000, obtenido %', v_plan; END IF;
  IF v_real <> 760  THEN RAISE EXCEPTION 'T7 FALLO: ton_real esperado 760, obtenido %', v_real; END IF;
  IF v_apego <> 76.00 THEN RAISE EXCEPTION 'T7 FALLO: apego esperado 76.00, obtenido %', v_apego; END IF;
  RAISE NOTICE 'OK  T7: plan vs real plan=% real=% apego=%', v_plan, v_real, v_apego;
END $$;

DO $$ DECLARE v_metros numeric; v_barr numeric; BEGIN
  SELECT sum(metros), sum(barrenos) INTO v_metros, v_barr
  FROM reportes.v_avance_barrenacion WHERE mina='La Cienega' AND fecha='2026-06-01';
  IF v_metros <> 20 THEN RAISE EXCEPTION 'T8 FALLO: metros esperado 20, obtenido %', v_metros; END IF;
  IF v_barr <> 10 THEN RAISE EXCEPTION 'T8 FALLO: barrenos esperado 10, obtenido %', v_barr; END IF;
  RAISE NOTICE 'OK  T8: avance barrenacion metros=% barrenos=%', v_metros, v_barr;
END $$;

DO $$ DECLARE v_dtr numeric; v_dpt numeric; BEGIN
  SELECT dif_tumbe_reserva, dif_planta_tumbe INTO v_dtr, v_dpt
  FROM reportes.v_reconciliacion WHERE segmento='SEG-2400W-001';
  IF v_dtr <> -50 THEN RAISE EXCEPTION 'T9 FALLO: dif tumbe-reserva esperado -50, obtenido %', v_dtr; END IF;
  IF v_dpt <> -50 THEN RAISE EXCEPTION 'T9 FALLO: dif planta-tumbe esperado -50, obtenido %', v_dpt; END IF;
  RAISE NOTICE 'OK  T9: reconciliacion dif(tumbe-reserva)=% dif(planta-tumbe)=%', v_dtr, v_dpt;
END $$;

DO $$ DECLARE v_op numeric; v_dem numeric; BEGIN
  SELECT horas_operativas, horas_demora INTO v_op, v_dem
  FROM reportes.v_disponibilidad_equipo WHERE equipo='EQ-001' AND fecha='2026-06-01';
  IF v_op <> 7.5 THEN RAISE EXCEPTION 'T10 FALLO: horas_operativas esperado 7.5, obtenido %', v_op; END IF;
  IF v_dem <> 1.0 THEN RAISE EXCEPTION 'T10 FALLO: horas_demora esperado 1.0, obtenido %', v_dem; END IF;
  RAISE NOTICE 'OK  T10: tiempos operativos camion op=%h demora=%h', v_op, v_dem;
END $$;

DO $$ DECLARE v_var numeric; BEGIN
  SELECT variacion INTO v_var
  FROM reportes.v_presupuesto_vs_real pr JOIN catalogos.mina m ON m.id=pr.id_mina
  WHERE m.nombre='La Cienega' AND pr.periodo='2026-06';
  IF v_var <> -5000 THEN RAISE EXCEPTION 'T11 FALLO: variacion esperado -5000, obtenido %', v_var; END IF;
  RAISE NOTICE 'OK  T11: presupuesto vs real variacion=% (real 45000 - ppto 50000)', v_var;
END $$;

-- ===== CAPACIDADES NUEVAS =====
DO $$ DECLARE v_ac numeric; v_mo numeric; v_dif numeric; BEGIN
  SELECT ton_acarreada, ton_molida, diferencia INTO v_ac, v_mo, v_dif
  FROM reportes.v_produccion_vs_molienda pm JOIN catalogos.mina m ON m.id=pm.id_mina
  WHERE m.nombre='La Cienega' AND pm.periodo='2026-06';
  IF v_ac <> 760 THEN RAISE EXCEPTION 'T17 FALLO: ton_acarreada esperado 760, obtenido %', v_ac; END IF;
  IF v_mo <> 720 THEN RAISE EXCEPTION 'T17 FALLO: ton_molida esperado 720, obtenido %', v_mo; END IF;
  RAISE NOTICE 'OK  T17: produccion vs molienda acarreado=% molido=% dif=%', v_ac, v_mo, v_dif;
END $$;

DO $$ DECLARE v_util numeric; v_meta numeric; BEGIN
  SELECT utilizacion_pct, utilizacion_meta_pct INTO v_util, v_meta
  FROM reportes.v_tiempos_vs_estandar WHERE equipo='EQ-001' AND fecha='2026-06-01';
  IF v_meta <> 80 THEN RAISE EXCEPTION 'T18 FALLO: meta utilizacion camion esperado 80, obtenido %', v_meta; END IF;
  IF v_util IS NULL THEN RAISE EXCEPTION 'T18 FALLO: utilizacion real nula'; END IF;
  RAISE NOTICE 'OK  T18: tiempos operativos vs estandar util_real=%%% meta=%%%', v_util, v_meta;
END $$;

DO $$ DECLARE v_rec numeric; BEGIN
  SELECT rec_au_pct INTO v_rec FROM reportes.v_balance_metalurgico WHERE codigo_lote='LOTE-2026-06-01';
  IF v_rec <> 92.5 THEN RAISE EXCEPTION 'T19 FALLO: recuperacion Au esperado 92.5, obtenido %', v_rec; END IF;
  RAISE NOTICE 'OK  T19: balance metalurgico recuperacion Au=%%%', v_rec;
END $$;

DO $$ DECLARE v_ton numeric; v_mts numeric; v_n int; BEGIN
  SELECT count(*) INTO v_n FROM reportes.v_obras_prioritarias WHERE obra='REB-2400W';
  IF v_n <> 1 THEN RAISE EXCEPTION 'T20 FALLO: REB-2400W deberia ser obra prioritaria'; END IF;
  SELECT ton_mineral, metros_barrenados INTO v_ton, v_mts FROM reportes.v_obras_prioritarias WHERE obra='REB-2400W';
  IF v_ton <> 760 THEN RAISE EXCEPTION 'T20 FALLO: ton_mineral obra esperado 760, obtenido %', v_ton; END IF;
  IF v_mts <> 20 THEN RAISE EXCEPTION 'T20 FALLO: metros obra esperado 20, obtenido %', v_mts; END IF;
  RAISE NOTICE 'OK  T20: obra prioritaria REB-2400W ton=% metros=%', v_ton, v_mts;
END $$;

DO $$ DECLARE v_inv numeric; BEGIN
  SELECT inversion INTO v_inv
  FROM reportes.v_inversion_obra WHERE obra='REB-2400W' AND periodo='2026-06';
  IF v_inv <> 120000 THEN RAISE EXCEPTION 'T21 FALLO: inversion REB esperado 120000, obtenido %', v_inv; END IF;
  RAISE NOTICE 'OK  T21: inversion por obra REB-2400W = %', v_inv;
END $$;

-- ===== GOBERNANZA: invariantes estructurales =====
DO $$ DECLARE v_falt int; BEGIN
  SELECT count(*) INTO v_falt FROM (
    SELECT table_schema, table_name FROM information_schema.tables
      WHERE table_schema IN ('catalogos','produccion','planeacion','reconciliacion','costos','beneficio','estandares','inversiones') AND table_type='BASE TABLE'
    EXCEPT
    SELECT table_schema, table_name FROM information_schema.columns
      WHERE column_name='id_empresa' AND table_schema IN ('catalogos','produccion','planeacion','reconciliacion','costos','beneficio','estandares','inversiones')
  ) f;
  IF v_falt <> 0 THEN RAISE EXCEPTION 'T12 FALLO: % tablas de negocio sin id_empresa', v_falt; END IF;
  RAISE NOTICE 'OK  T12: todas las tablas de negocio tienen id_empresa (multi-tenant)';
END $$;

DO $$ DECLARE v_bad int; BEGIN
  SELECT count(*) INTO v_bad FROM information_schema.columns
  WHERE table_schema='produccion'
    AND table_name IN ('acarreo_viaje','rezagado_ciclo','barrenacion_avance','barrenacion_ejecutado','consumo_explosivo','demora_equipo')
    AND column_name IN ('eliminado_en','actualizado_en');
  IF v_bad <> 0 THEN RAISE EXCEPTION 'T13 FALLO: % columnas de mutacion en tablas evento', v_bad; END IF;
  RAISE NOTICE 'OK  T13: tablas evento son append-only (sin actualizado/eliminado_en)';
END $$;

DO $$ DECLARE v_n int; BEGIN
  SELECT count(*) INTO v_n FROM produccion.acarreo_viaje;
  IF v_n <> 4 THEN RAISE EXCEPTION 'T14 FALLO: se esperaban 4 viajes de acarreo, hay %', v_n; END IF;
  RAISE NOTICE 'OK  T14: integridad seed - 4 viajes de acarreo';
END $$;

DO $$ DECLARE v_emp uuid; v_id1 uuid; v_id2 uuid; BEGIN
  SELECT id INTO v_emp FROM gobierno.empresa WHERE codigo='MIN';
  INSERT INTO catalogos.obra (id_empresa, id_mina, codigo) VALUES (v_emp,(SELECT id FROM catalogos.mina WHERE id_empresa=v_emp LIMIT 1),'TMP_REALTA') RETURNING id INTO v_id1;
  BEGIN
    INSERT INTO catalogos.obra (id_empresa, id_mina, codigo) VALUES (v_emp,(SELECT id_mina FROM catalogos.obra WHERE id=v_id1),'TMP_REALTA');
    RAISE EXCEPTION 'T15 FALLO: permitio dos obras activas con mismo codigo';
  EXCEPTION WHEN unique_violation THEN NULL;
  END;
  UPDATE catalogos.obra SET eliminado_en = now() WHERE id = v_id1;
  INSERT INTO catalogos.obra (id_empresa, id_mina, codigo) VALUES (v_emp,(SELECT id FROM catalogos.mina WHERE id_empresa=v_emp LIMIT 1),'TMP_REALTA') RETURNING id INTO v_id2;
  DELETE FROM catalogos.obra WHERE id IN (v_id1, v_id2);
  RAISE NOTICE 'OK  T15: indice unico parcial - rechaza activos duplicados y permite re-alta tras baja logica';
END $$;

DO $$ DECLARE v_falt int; BEGIN
  SELECT count(*) INTO v_falt FROM (
    SELECT table_schema, table_name FROM information_schema.columns
      WHERE column_name='eliminado_en' AND table_schema IN ('catalogos','produccion','planeacion','reconciliacion','costos','beneficio','estandares','inversiones','gobierno')
    EXCEPT
    SELECT table_schema, table_name FROM information_schema.columns
      WHERE column_name='actualizado_por_usuario_id' AND table_schema IN ('catalogos','produccion','planeacion','reconciliacion','costos','beneficio','estandares','inversiones','gobierno')
  ) f;
  IF v_falt <> 0 THEN RAISE EXCEPTION 'T16 FALLO: % tablas con eliminado_en pero sin auditoria de actualizacion', v_falt; END IF;
  RAISE NOTICE 'OK  T16: trazabilidad consistente (borrado logico => auditoria completa)';
END $$;

-- T22: productividad de barrenación SIN fan-out (regresión del bug detectado por auditoría)
DO $$ DECLARE v_h numeric; v_m numeric; v_r numeric; BEGIN
  SELECT horas_percusion, metros, metros_por_hora INTO v_h, v_m, v_r
  FROM reportes.v_productividad_barrenacion WHERE equipo='EQ-004' AND fecha='2026-06-01';
  IF v_h <> 8.5 THEN RAISE EXCEPTION 'T22 FALLO: horas_percusion esperado 8.5 (sin fan-out), obtenido %', v_h; END IF;
  IF v_m <> 20 THEN RAISE EXCEPTION 'T22 FALLO: metros esperado 20, obtenido %', v_m; END IF;
  IF v_r <> 2.35 THEN RAISE EXCEPTION 'T22 FALLO: metros_por_hora esperado 2.35, obtenido %', v_r; END IF;
  RAISE NOTICE 'OK  T22: productividad barrenacion horas=% metros=% ratio=% (sin fan-out)', v_h, v_m, v_r;
END $$;

-- T23: "Ejecutado" (lista fija múltiple) — secciones de barreno por avance
DO $$ DECLARE v_n int; BEGIN
  SELECT count(*) INTO v_n FROM produccion.barrenacion_ejecutado be
  JOIN produccion.barrenacion_avance ba ON ba.id = be.id_barrenacion_avance
  WHERE ba.lugar='Tajo 2400W Bl-5';
  IF v_n <> 4 THEN RAISE EXCEPTION 'T23 FALLO: tipos de barreno ejecutados esperado 4, obtenido %', v_n; END IF;
  RAISE NOTICE 'OK  T23: barrenacion_ejecutado (lista fija multiple) = % secciones en Bl-5', v_n;
END $$;

-- ===== AISLAMIENTO MULTI-TENANT EN LA BD (RLS + FKs compuestas + privilegios) =====
-- Los ids de ambos tenants se capturan como variables psql ANTES de bajar
-- privilegios, y se publican como GUCs para usarlos dentro de los DO blocks.
SELECT id AS emp_a FROM gobierno.empresa WHERE codigo='MIN' \gset
SELECT id AS emp_b FROM gobierno.empresa WHERE codigo='MIN2' \gset
SELECT id AS mina_b FROM catalogos.mina WHERE nombre='Mina Norte B' \gset
SELECT set_config('app.t_emp_b',  :'emp_b',  false);
SELECT set_config('app.t_mina_b', :'mina_b', false);

-- T24: RLS aísla la lectura por tenant
SELECT set_config('app.empresa_actual', :'emp_a', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM catalogos.mina;
  IF v <> 3 THEN RAISE EXCEPTION 'T24 FALLO: tenant A debe ver 3 minas, ve %', v; END IF;
  SELECT count(*) INTO v FROM gobierno.empresa;
  IF v <> 1 THEN RAISE EXCEPTION 'T24 FALLO: el tenant debe verse solo a si mismo, ve % empresas', v; END IF;
END $$;
RESET ROLE;
SELECT set_config('app.empresa_actual', :'emp_b', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM catalogos.mina;
  IF v <> 1 THEN RAISE EXCEPTION 'T24 FALLO: tenant B debe ver 1 mina, ve %', v; END IF;
  SELECT count(*) INTO v FROM produccion.acarreo_viaje;
  IF v <> 0 THEN RAISE EXCEPTION 'T24 FALLO: tenant B no tiene viajes, ve %', v; END IF;
  RAISE NOTICE 'OK  T24: RLS aisla por tenant (A=3 minas y solo su empresa; B=1 mina y 0 viajes)';
END $$;
RESET ROLE;

-- T25: fail-closed — sin tenant fijado, la BD no devuelve nada
SELECT set_config('app.empresa_actual', '', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM catalogos.mina;
  IF v <> 0 THEN RAISE EXCEPTION 'T25 FALLO: sin app.empresa_actual deben verse 0 filas, se ven %', v; END IF;
  RAISE NOTICE 'OK  T25: fail-closed - sin tenant fijado la BD devuelve 0 filas';
END $$;
RESET ROLE;

-- T26: WITH CHECK — la app (con tenant A) no puede escribir filas de B
SELECT set_config('app.empresa_actual', :'emp_a', false);
SET ROLE aplicacion;
DO $$ BEGIN
  BEGIN
    INSERT INTO catalogos.obra (id_empresa, id_mina, codigo)
    VALUES (current_setting('app.t_emp_b')::uuid, current_setting('app.t_mina_b')::uuid, 'OBRA-CRUZADA');
    RAISE EXCEPTION 'T26 FALLO: INSERT hacia otro tenant fue aceptado';
  EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'OK  T26: RLS (WITH CHECK) bloquea INSERT hacia otro tenant';
  END;
END $$;
RESET ROLE;

-- T27: FK compuesta — ni siquiera SIN RLS se puede referenciar el catálogo de otro tenant
DO $$ DECLARE v_tipo uuid; v_mod uuid; BEGIN
  SELECT te.id INTO v_tipo FROM catalogos.tipo_equipo te JOIN gobierno.empresa e ON e.id=te.id_empresa WHERE e.codigo='MIN' AND te.codigo='CAMION';
  SELECT mt.id INTO v_mod  FROM catalogos.modulo_trabajo mt JOIN gobierno.empresa e ON e.id=mt.id_empresa WHERE e.codigo='MIN' AND mt.codigo='ACARREO';
  BEGIN
    INSERT INTO catalogos.equipo (id_empresa, id_mina, id_tipo_equipo, id_modulo_trabajo, codigo)
    VALUES (current_setting('app.t_emp_b')::uuid, current_setting('app.t_mina_b')::uuid, v_tipo, v_mod, 'EQ-CRUZADO');
    RAISE EXCEPTION 'T27 FALLO: la FK compuesta permitio referenciar el catalogo de otro tenant';
  EXCEPTION WHEN foreign_key_violation THEN
    RAISE NOTICE 'OK  T27: FK compuesta (id_empresa,id) hace imposible el cruce de tenants';
  END;
END $$;

-- T28: privilegios — eventos sin UPDATE (append-only) y sin DELETE físico para la app
SELECT set_config('app.empresa_actual', :'emp_a', false);
SET ROLE aplicacion;
DO $$ BEGIN
  BEGIN
    UPDATE produccion.acarreo_viaje SET toneladas = toneladas;
    RAISE EXCEPTION 'T28 FALLO: UPDATE sobre evento append-only fue permitido';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
  BEGIN
    DELETE FROM catalogos.mina;
    RAISE EXCEPTION 'T28 FALLO: DELETE fisico fue permitido a la app';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
  RAISE NOTICE 'OK  T28: privilegios refuerzan append-only (eventos sin UPDATE) y borrado solo logico (sin DELETE)';
END $$;
RESET ROLE;

-- T29: vistas y materializada respetan el tenant
SELECT set_config('app.empresa_actual', :'emp_a', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM reportes.v_produccion_acarreo;
  IF v = 0 THEN RAISE EXCEPTION 'T29 FALLO: tenant A debe ver su produccion en la vista'; END IF;
  SELECT count(*) INTO v FROM reportes.v_plan_vs_real_mensual;
  IF v <> 1 THEN RAISE EXCEPTION 'T29 FALLO: tenant A debe ver 1 fila de plan vs real, ve %', v; END IF;
  BEGIN
    SELECT count(*) INTO v FROM reportes.mv_plan_vs_real_mensual;
    RAISE EXCEPTION 'T29 FALLO: la materializada quedo expuesta directamente a la app';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
END $$;
RESET ROLE;
SELECT set_config('app.empresa_actual', :'emp_b', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM reportes.v_produccion_acarreo;
  IF v <> 0 THEN RAISE EXCEPTION 'T29 FALLO: tenant B no tiene produccion y ve % filas', v; END IF;
  SELECT count(*) INTO v FROM reportes.v_plan_vs_real_mensual;
  IF v <> 0 THEN RAISE EXCEPTION 'T29 FALLO: tenant B no tiene plan y ve % filas', v; END IF;
  RAISE NOTICE 'OK  T29: vistas (security_invoker) y materializada (via vista filtrada) respetan el tenant';
END $$;
RESET ROLE;

-- T30: RBAC aislado por tenant — cada empresa ve SOLO sus roles y asignaciones
SELECT set_config('app.empresa_actual', :'emp_a', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM gobierno.rol;
  IF v <> 6 THEN RAISE EXCEPTION 'T30 FALLO: tenant A debe ver 6 roles de sistema, ve %', v; END IF;
  SELECT count(*) INTO v FROM gobierno.usuario_rol ur
    JOIN gobierno.usuario u ON u.id = ur.id_usuario
    WHERE u.usuario = 'admin.mina' AND ur.eliminado_en IS NULL;
  IF v <> 1 THEN RAISE EXCEPTION 'T30 FALLO: admin.mina debe tener 1 rol vigente, tiene %', v; END IF;
  RAISE NOTICE 'OK  T30: RBAC por tenant (6 roles sistema; admin.mina con ADMIN_EMPRESA vigente)';
END $$;
RESET ROLE;

-- T31: superadmin invisible para la app; el rol plataforma administra TODOS los tenants
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  BEGIN
    SELECT count(*) INTO v FROM gobierno.superadmin;
    RAISE EXCEPTION 'T31 FALLO: superadmin quedo visible para aplicacion';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
END $$;
RESET ROLE;
SET ROLE plataforma;
DO $$ DECLARE v_emp int; v_sa int; BEGIN
  SELECT count(*) INTO v_emp FROM gobierno.empresa;
  IF v_emp < 2 THEN RAISE EXCEPTION 'T31 FALLO: plataforma debe ver todas las empresas, ve %', v_emp; END IF;
  SELECT count(*) INTO v_sa FROM gobierno.superadmin;
  IF v_sa <> 1 THEN RAISE EXCEPTION 'T31 FALLO: plataforma debe ver 1 superadmin, ve %', v_sa; END IF;
  RAISE NOTICE 'OK  T31: superadmin oculto a la app; plataforma ve % empresas y gestiona superadmins', v_emp;
END $$;
RESET ROLE;

-- T32: FK compuesta en RBAC — imposible asignar a un usuario un rol de OTRA empresa
DO $$ DECLARE v_rol_b uuid; BEGIN
  SELECT r.id INTO v_rol_b FROM gobierno.rol r
    JOIN gobierno.empresa e ON e.id = r.id_empresa
    WHERE e.codigo = 'MIN2' AND r.codigo = 'ADMIN_EMPRESA';
  BEGIN
    INSERT INTO gobierno.usuario_rol (id_empresa, id_usuario, id_rol)
    VALUES ((SELECT id FROM gobierno.empresa WHERE codigo='MIN'),
            (SELECT id FROM gobierno.usuario WHERE usuario='admin.mina'),
            v_rol_b);
    RAISE EXCEPTION 'T32 FALLO: se asigno un rol de otra empresa';
  EXCEPTION WHEN foreign_key_violation THEN
    RAISE NOTICE 'OK  T32: FK compuesta impide asignar roles de otra empresa';
  END;
END $$;

-- T33: catálogo de permisos GLOBAL — legible por los tenants, escribible solo por plataforma
SELECT set_config('app.empresa_actual', :'emp_a', false);
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM gobierno.permiso WHERE eliminado_en IS NULL;
  IF v <> 29 THEN RAISE EXCEPTION 'T33 FALLO: catalogo esperado 29 permisos, hay %', v; END IF;
  BEGIN
    INSERT INTO gobierno.permiso (codigo, descripcion, modulo) VALUES ('hack.crear','x','gobierno');
    RAISE EXCEPTION 'T33 FALLO: un tenant pudo crear permisos';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
  RAISE NOTICE 'OK  T33: catalogo global de 29 permisos, solo-lectura para los tenants';
END $$;
RESET ROLE;

-- T34: ADMIN_EMPRESA tiene TODOS los permisos; la matriz está aislada por tenant
SET ROLE aplicacion;
DO $$ DECLARE v int; v_emp int; BEGIN
  SELECT count(*) INTO v FROM gobierno.rol_permiso rp
    JOIN gobierno.rol r ON r.id = rp.id_rol
    WHERE r.codigo='ADMIN_EMPRESA' AND rp.eliminado_en IS NULL;
  IF v <> 29 THEN RAISE EXCEPTION 'T34 FALLO: ADMIN_EMPRESA debe tener 29 permisos, tiene %', v; END IF;
  SELECT count(DISTINCT id_empresa) INTO v_emp FROM gobierno.rol_permiso;
  IF v_emp <> 1 THEN RAISE EXCEPTION 'T34 FALLO: la matriz muestra % empresas (RLS roto)', v_emp; END IF;
  RAISE NOTICE 'OK  T34: ADMIN_EMPRESA con los 29 permisos; matriz aislada por tenant';
END $$;
RESET ROLE;

-- T35: roles de SISTEMA protegidos; roles propios personalizables
SET ROLE aplicacion;
DO $$ DECLARE v int; v_rol uuid; BEGIN
  UPDATE gobierno.rol SET descripcion = 'hackeado' WHERE codigo='OPERADOR';
  GET DIAGNOSTICS v = ROW_COUNT;
  IF v <> 0 THEN RAISE EXCEPTION 'T35 FALLO: se pudo editar un rol de sistema (% filas)', v; END IF;
  BEGIN
    INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
    SELECT r.id_empresa, r.id, p.id FROM gobierno.rol r, gobierno.permiso p
    WHERE r.codigo='OPERADOR' AND p.codigo='auditoria.ver';
    RAISE EXCEPTION 'T35 FALLO: se pudo alterar la matriz de un rol de sistema';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
  INSERT INTO gobierno.rol (id_empresa, codigo, descripcion)
  VALUES (current_setting('app.empresa_actual')::uuid, 'SUPERVISOR_PATIO', 'Rol propio de la empresa')
  ON CONFLICT (id_empresa, codigo) WHERE eliminado_en IS NULL DO NOTHING;
  SELECT id INTO v_rol FROM gobierno.rol WHERE codigo='SUPERVISOR_PATIO' AND eliminado_en IS NULL;
  INSERT INTO gobierno.rol_permiso (id_empresa, id_rol, id_permiso)
  SELECT current_setting('app.empresa_actual')::uuid, v_rol, p.id FROM gobierno.permiso p WHERE p.codigo='reportes.ver'
  ON CONFLICT (id_empresa, id_rol, id_permiso) WHERE eliminado_en IS NULL DO NOTHING;
  RAISE NOTICE 'OK  T35: roles de sistema intocables; rol propio creado y personalizado por el admin';
END $$;
RESET ROLE;

-- T36: branding/configuración de empresa (logo, color, zona horaria, moneda)
DO $$ DECLARE v text; BEGIN
  SELECT zona_horaria INTO v FROM gobierno.empresa WHERE codigo='MIN';
  IF v <> 'America/Mexico_City' THEN RAISE EXCEPTION 'T36 FALLO: zona_horaria default, obtenido %', v; END IF;
END $$;
SET ROLE plataforma;
DO $$ BEGIN
  BEGIN
    UPDATE gobierno.empresa SET color_primario = 'naranja' WHERE codigo='MIN';
    RAISE EXCEPTION 'T36 FALLO: color invalido aceptado';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  UPDATE gobierno.empresa
     SET logo_url = 'https://cdn.ejemplo.com/min/logo.png', color_primario = '#7A4B16'
   WHERE codigo='MIN';
END $$;
RESET ROLE;
SET ROLE aplicacion;
DO $$ DECLARE v text; BEGIN
  SELECT logo_url INTO v FROM gobierno.empresa;
  IF v <> 'https://cdn.ejemplo.com/min/logo.png' THEN RAISE EXCEPTION 'T36 FALLO: el tenant no ve su logo (%)', v; END IF;
  RAISE NOTICE 'OK  T36: branding por empresa (logo + color validado + zona horaria/moneda default)';
END $$;
RESET ROLE;

-- T37: observaciones en los partes (requerimiento de operación real)
SET ROLE aplicacion;
DO $$ DECLARE v int; v_id uuid; v_txt text; BEGIN
  SELECT count(*) INTO v FROM information_schema.columns
   WHERE table_schema='produccion' AND column_name='observaciones'
     AND table_name IN ('parte_acarreo','parte_rezagado','parte_barrenacion');
  IF v <> 3 THEN RAISE EXCEPTION 'T37 FALLO: observaciones presente en % de 3 partes', v; END IF;
  SELECT id INTO v_id FROM produccion.parte_acarreo LIMIT 1;
  UPDATE produccion.parte_acarreo SET observaciones = 'Turno sin novedades; tope 2400W con agua.' WHERE id = v_id;
  SELECT observaciones INTO v_txt FROM produccion.parte_acarreo WHERE id = v_id;
  IF v_txt IS NULL THEN RAISE EXCEPTION 'T37 FALLO: observaciones no persistio'; END IF;
  RAISE NOTICE 'OK  T37: observaciones en los 3 partes, escribible por la app';
END $$;
RESET ROLE;

-- T38: permisos efectivos (v_permisos_usuario) — lo que el backend pone en el JWT
SET ROLE aplicacion;
DO $$ DECLARE v int; BEGIN
  SELECT count(*) INTO v FROM gobierno.v_permisos_usuario WHERE usuario='admin.mina';
  IF v < 29 THEN RAISE EXCEPTION 'T38 FALLO: admin.mina debe tener >=29 permisos efectivos, tiene %', v; END IF;
  IF NOT EXISTS (SELECT 1 FROM gobierno.v_permisos_usuario WHERE usuario='admin.mina' AND permiso='usuarios.crear') THEN
    RAISE EXCEPTION 'T38 FALLO: admin.mina sin usuarios.crear';
  END IF;
  RAISE NOTICE 'OK  T38: v_permisos_usuario entrega los permisos efectivos (admin.mina con % permisos)', v;
END $$;
RESET ROLE;

SELECT set_config('app.empresa_actual', '', false);

DO $$ BEGIN RAISE NOTICE '========================================'; RAISE NOTICE 'TODOS LOS TESTS PASARON'; RAISE NOTICE '========================================'; END $$;
