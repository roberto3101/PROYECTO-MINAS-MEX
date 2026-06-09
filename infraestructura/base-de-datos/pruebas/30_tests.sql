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
    VALUES ((SELECT id FROM gobierno.empresa LIMIT 1), '00000000-0000-0000-0000-000000000000',
            (SELECT id FROM catalogos.tipo_equipo LIMIT 1), (SELECT id FROM catalogos.modulo_trabajo LIMIT 1), 'EQ-FK');
    RAISE EXCEPTION 'T4 FALLO: FK a mina inexistente fue aceptada';
  EXCEPTION WHEN foreign_key_violation THEN RAISE NOTICE 'OK  T4: FK a mina inexistente rechazada';
  END;
END $$;

DO $$ BEGIN
  BEGIN
    INSERT INTO catalogos.mina (id_empresa, nombre, area)
    VALUES ((SELECT id FROM gobierno.empresa LIMIT 1), 'La Cienega', 'Cienega');
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
  SELECT id INTO v_emp FROM gobierno.empresa LIMIT 1;
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

DO $$ BEGIN RAISE NOTICE '========================================'; RAISE NOTICE 'TODOS LOS TESTS PASARON'; RAISE NOTICE '========================================'; END $$;
