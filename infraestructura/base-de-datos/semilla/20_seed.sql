-- ============================================================
-- 20 · Semilla — datos de ejemplo (basados en el Excel real)
-- ============================================================
-- IDs fijos para entidades clave (determinístico, testeable).

\set emp        '11111111-1111-1111-1111-111111111111'
\set usr        '22222222-2222-2222-2222-222222222222'
\set empB       '99999999-9999-9999-9999-999999999999'
\set m_b        'aaaa0000-0000-0000-0000-0000000000b1'
\set m_cien     'aaaa0000-0000-0000-0000-000000000001'
\set m_esp      'aaaa0000-0000-0000-0000-000000000002'
\set m_sat      'aaaa0000-0000-0000-0000-000000000003'
\set eq_camion  'bbbb0000-0000-0000-0000-000000000001'
\set eq_scoop   'bbbb0000-0000-0000-0000-000000000002'
\set eq_jum_lin 'bbbb0000-0000-0000-0000-000000000003'
\set eq_jum_lg  'bbbb0000-0000-0000-0000-000000000004'
\set eq_ancla   'bbbb0000-0000-0000-0000-000000000005'
\set e_juan     'cccc0000-0000-0000-0000-000000000001'
\set e_cesar    'cccc0000-0000-0000-0000-000000000002'
\set e_pedro    'cccc0000-0000-0000-0000-000000000003'
\set e_maria    'cccc0000-0000-0000-0000-000000000004'
\set e_rey      'cccc0000-0000-0000-0000-000000000005'
\set obra_reb   'f6660000-0000-0000-0000-000000000001'
\set obra_rpa   'f6660000-0000-0000-0000-000000000002'
\set plan_cuts  'dddd0000-0000-0000-0000-000000000001'
\set reb_tase   'dddd0000-0000-0000-0000-000000000010'
\set bp_819     'dddd0000-0000-0000-0000-000000000100'
\set seg_001    'eeee0000-0000-0000-0000-000000000001'
\set pa1        'f1110000-0000-0000-0000-000000000001'
\set pa2        'f1110000-0000-0000-0000-000000000002'
\set pr1        'f2220000-0000-0000-0000-000000000001'
\set pb1        'f3330000-0000-0000-0000-000000000001'
\set contra1    'f5550000-0000-0000-0000-000000000001'
\set est1       'f4440000-0000-0000-0000-000000000001'
\set lote1      'f7770000-0000-0000-0000-000000000001'

-- ---- Gobierno ----
INSERT INTO gobierno.empresa (id, codigo, razon_social) VALUES (:'emp','MIN','Minera Ejemplo SA');
INSERT INTO gobierno.usuario (id, id_empresa, usuario, nombre) VALUES (:'usr', :'emp', 'seed', 'Carga Inicial');

-- ---- Tenant B mínimo (verifica aislamiento RLS y FKs compuestas) ----
INSERT INTO gobierno.empresa (id, codigo, razon_social) VALUES (:'empB','MIN2','Minera Dos SA');
INSERT INTO catalogos.mina (id, id_empresa, nombre, area) VALUES (:'m_b', :'empB','Mina Norte B','Norte');

-- ---- Lookups ----
INSERT INTO catalogos.tipo_obra (id_empresa, codigo, descripcion) VALUES
 (:'emp','RPA','Rampa'),(:'emp','LAT','Lateral'),(:'emp','ACG','Acceso a galería'),
 (:'emp','CRO','Crucero'),(:'emp','REB','Rebaje'),(:'emp','STOCK','Stock'),(:'emp','STILL','Still');
INSERT INTO catalogos.sistema_minado (id_empresa, codigo, descripcion) VALUES
 (:'emp','BL','Bench and Long'),(:'emp','CR','Corte y relleno'),(:'emp','TSC','Tumbe sobre carga');
INSERT INTO catalogos.tipo_equipo (id_empresa, codigo, descripcion) VALUES
 (:'emp','CAMION','Camion'),(:'emp','SCOOP','Scooptram'),(:'emp','JUMBO_LINEAL','Jumbo lineal'),
 (:'emp','JUMBO_LARGA','Jumbo barrenacion larga'),(:'emp','ANCLADOR','Anclador'),
 (:'emp','LOCOMOTORA','Locomotora'),(:'emp','TRACTOR','Tractor');
INSERT INTO catalogos.modulo_trabajo (id_empresa, codigo, descripcion) VALUES
 (:'emp','ACARREO','Acarreo'),(:'emp','REZAGADO','Rezagado'),(:'emp','BARRENACION','Barrenacion'),
 (:'emp','ANCLAJE','Anclaje'),(:'emp','RELLENO','Relleno');
INSERT INTO catalogos.tipo_mineral (id_empresa, codigo, descripcion, es_mineral) VALUES
 (:'emp','MINERAL','Mineral', true),(:'emp','NO_MINERAL','No mineral', false),(:'emp','TEPETATE','Tepetate', false);
INSERT INTO catalogos.departamento (id_empresa, codigo, descripcion) VALUES
 (:'emp','MINA','Mina'),(:'emp','PLANEACION','Planeacion'),(:'emp','GERENCIA','Gerencia');
INSERT INTO catalogos.puesto (id_empresa, codigo, descripcion) VALUES
 (:'emp','CAP_MINA','Capitan de Mina'),(:'emp','SUPERVISOR','Supervisor'),(:'emp','OP_JUMBO','Operador Jumbo'),
 (:'emp','OP_CAMION','Operador Camion'),(:'emp','GERENTE','Gerente Mina');
INSERT INTO catalogos.actividad (id_empresa, codigo, descripcion) VALUES
 (:'emp','DESARROLLO','Desarrollo'),(:'emp','TUMBE','Tumbe'),(:'emp','BARRENACION','Barrenacion'),
 (:'emp','REZAGADO','Rezagado'),(:'emp','ACARREO','Acarreo'),(:'emp','ANCLAJE','Anclaje');
INSERT INTO catalogos.tipo_demora (id_empresa, codigo, descripcion, es_programada) VALUES
 (:'emp','MECANICA','Falla mecanica', false),(:'emp','ELECTRICA','Falla electrica', false),
 (:'emp','OPERATIVA','Demora operativa', false),(:'emp','PROGRAMADA','Mantenimiento programado', true);
INSERT INTO catalogos.tipo_explosivo (id_empresa, codigo, descripcion, familia, unidad_medida) VALUES
 (:'emp','ANFO','Anfo','AGENTE','kg'),(:'emp','EMULSION','Emulsion','AGENTE','kg'),
 (:'emp','FULMINANTE','Fulminante','INICIADOR','pza'),(:'emp','CORDON','Cordon detonante','CORDON','m');
INSERT INTO catalogos.tipo_barreno (id_empresa, codigo, descripcion) VALUES
 (:'emp','CUELE','Cuele (arranque)'),(:'emp','CONTRACUELE','Contracuele'),(:'emp','DESTROZA','Destroza / ensanche'),
 (:'emp','CONTORNO','Contorno / recorte'),(:'emp','ZAPATERA','Zapatera / piso (arrastre)'),
 (:'emp','CORONA','Corona / hastiales'),(:'emp','AUXILIAR','Auxiliar');

-- ---- Minas ----
INSERT INTO catalogos.mina (id, id_empresa, nombre, area, niveles, densidad_promedio, seccion_obra_largo, seccion_obra_ancho) VALUES
 (:'m_cien', :'emp','La Cienega','Cienega','1900-2400',2.650,5,4),
 (:'m_esp',  :'emp','La Esperanza','Esperanza','2100-2550',2.700,4.5,3.5),
 (:'m_sat',  :'emp','Satelites Norte','Satelites','2000-2300',2.680,5,4);

-- ---- Equipos ----
INSERT INTO catalogos.equipo (id, id_empresa, id_mina, id_tipo_equipo, id_modulo_trabajo, codigo, fabricante, modelo, numero_serie, anio_fabricacion) VALUES
 (:'eq_camion', :'emp', :'m_cien', (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='CAMION'),      (SELECT id FROM catalogos.modulo_trabajo WHERE id_empresa=:'emp' AND codigo='ACARREO'),    'EQ-001','Sandvik','TH663','SN-001-2019',2019),
 (:'eq_scoop',  :'emp', :'m_cien', (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='SCOOP'),       (SELECT id FROM catalogos.modulo_trabajo WHERE id_empresa=:'emp' AND codigo='REZAGADO'),   'EQ-002','Atlas Copco','ST14','SN-002-2020',2020),
 (:'eq_jum_lin',:'emp', :'m_esp',  (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='JUMBO_LINEAL'),(SELECT id FROM catalogos.modulo_trabajo WHERE id_empresa=:'emp' AND codigo='BARRENACION'),'EQ-003','Sandvik','DL421','SN-003-2018',2018),
 (:'eq_jum_lg', :'emp', :'m_sat',  (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='JUMBO_LARGA'), (SELECT id FROM catalogos.modulo_trabajo WHERE id_empresa=:'emp' AND codigo='BARRENACION'),'EQ-004','Epiroc','L2C','SN-004-2021',2021),
 (:'eq_ancla',  :'emp', :'m_cien', (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='ANCLADOR'),    (SELECT id FROM catalogos.modulo_trabajo WHERE id_empresa=:'emp' AND codigo='BARRENACION'),'EQ-005','Atlas Copco','Boltec E','SN-005-2022',2022);

-- ---- Empleados ----
INSERT INTO catalogos.empleado (id, id_empresa, id_mina, numero_nomina, nombre_completo, id_departamento, id_puesto, centro_costo) VALUES
 (:'e_juan', :'emp', :'m_cien','EMP001','JUAN FCO. RODRIGUEZ V.',(SELECT id FROM catalogos.departamento WHERE id_empresa=:'emp' AND codigo='MINA'),     (SELECT id FROM catalogos.puesto WHERE id_empresa=:'emp' AND codigo='CAP_MINA'),'007-0001'),
 (:'e_cesar',:'emp', :'m_cien','EMP002','CESAR HUERTA',          (SELECT id FROM catalogos.departamento WHERE id_empresa=:'emp' AND codigo='PLANEACION'),(SELECT id FROM catalogos.puesto WHERE id_empresa=:'emp' AND codigo='SUPERVISOR'),'007-0002'),
 (:'e_pedro',:'emp', :'m_esp', 'EMP003','PEDRO LEAL GOMEZ',      (SELECT id FROM catalogos.departamento WHERE id_empresa=:'emp' AND codigo='MINA'),     (SELECT id FROM catalogos.puesto WHERE id_empresa=:'emp' AND codigo='OP_JUMBO'),'007-0003'),
 (:'e_maria',:'emp', :'m_sat', 'EMP004','MARIA LOPEZ RUIZ',      (SELECT id FROM catalogos.departamento WHERE id_empresa=:'emp' AND codigo='MINA'),     (SELECT id FROM catalogos.puesto WHERE id_empresa=:'emp' AND codigo='OP_CAMION'),'007-0004'),
 (:'e_rey',  :'emp', :'m_cien','EMP005','REYNALDO JIMENEZ',      (SELECT id FROM catalogos.departamento WHERE id_empresa=:'emp' AND codigo='GERENCIA'), (SELECT id FROM catalogos.puesto WHERE id_empresa=:'emp' AND codigo='GERENTE'),'001-0001');

-- ---- Minerales ----
INSERT INTO catalogos.mineral (id_empresa, id_mina, id_tipo_mineral, nombre, unidad_medida, ley) VALUES
 (:'emp', :'m_cien', (SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),  'Plata','g/t',350.5),
 (:'emp', :'m_cien', (SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),  'Oro','g/t',2.8),
 (:'emp', :'m_cien', (SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='TEPETATE'), 'Tepetate','ton',0),
 (:'emp', :'m_esp',  (SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),  'Plomo','g/t',120),
 (:'emp', :'m_esp',  (SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),  'Zinc','g/t',85.5);

-- ---- Obras (entidad) ----
INSERT INTO catalogos.obra (id, id_empresa, id_mina, id_tipo_obra, codigo, nombre, ubicacion, es_prioritaria) VALUES
 (:'obra_reb', :'emp', :'m_cien', (SELECT id FROM catalogos.tipo_obra WHERE id_empresa=:'emp' AND codigo='REB'),'REB-2400W','Rebaje 2400 Oeste','Nivel 2400', true),
 (:'obra_rpa', :'emp', :'m_cien', (SELECT id FROM catalogos.tipo_obra WHERE id_empresa=:'emp' AND codigo='RPA'),'RPA-2140','Rampa 2140','Nivel 2140', false);

-- ---- Plan (escenario CUTS2026) ----
INSERT INTO planeacion.rebaje (id, id_empresa, id_mina, id_obra, codigo, descripcion, id_tipo_obra, id_sistema_minado) VALUES
 (:'reb_tase', :'emp', :'m_cien', :'obra_reb','TASE2040E','Rebaje 2400W',(SELECT id FROM catalogos.tipo_obra WHERE id_empresa=:'emp' AND codigo='REB'),(SELECT id FROM catalogos.sistema_minado WHERE id_empresa=:'emp' AND codigo='BL'));
INSERT INTO planeacion.plan (id, id_empresa, id_mina, nombre, anio, t_inversion, enfoque) VALUES
 (:'plan_cuts', :'emp', :'m_cien','CUTS2026',2026,'OPEX','BASE');
INSERT INTO planeacion.bloque_programado (id, id_empresa, id_plan, id_rebaje, codigo_bloque, fecha_inicio, largo, ancho_veta, alto, fraccion, waste, minado, factor_dilucion, dilucion_pct, toneladas, toneladas_diluidas, ley_au, ley_ag, ley_pb, ley_zn, nsr, clasif_h, corte_h) VALUES
 (:'bp_819', :'emp', :'plan_cuts', :'reb_tase','819','2026-06-01',10,1.14,2,0.08,1.22,1.070,1.070,7,146.67,143.329,0.834,1.326,3,4,0,'Opex',23);
INSERT INTO planeacion.meta_periodo (id_empresa, id_bloque_programado, horizonte, periodo_etiqueta, toneladas_plan) VALUES
 (:'emp', :'bp_819','ANUAL','2026',12000),
 (:'emp', :'bp_819','MENSUAL','2026-06',1000);

-- ---- Producción: ACARREO ----
INSERT INTO produccion.parte_acarreo (id, id_empresa, id_mina, id_obra, id_equipo, id_operador, id_supervisor, fecha, turno, horometro_inicial, horometro_final) VALUES
 (:'pa1', :'emp', :'m_cien', :'obra_reb', :'eq_camion', :'e_maria', :'e_cesar','2026-06-01','M',5420.5,5428),
 (:'pa2', :'emp', :'m_cien', :'obra_reb', :'eq_camion', :'e_maria', :'e_cesar','2026-06-02','V',5428,5435);
INSERT INTO produccion.acarreo_viaje (id_empresa, id_parte_acarreo, desde, hasta, id_tipo_mineral, toneladas) VALUES
 (:'emp', :'pa1','Tajo 2400W','Patio de Mineral',(SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),400),
 (:'emp', :'pa1','Tajo 2400W','Silo',(SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='TEPETATE'),120),
 (:'emp', :'pa2','Tajo 2400W','Patio de Mineral',(SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),360),
 (:'emp', :'pa2','Tajo 2400W','Silo',(SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='TEPETATE'),80);

-- ---- Producción: REZAGADO ----
INSERT INTO produccion.parte_rezagado (id, id_empresa, id_mina, id_obra, id_equipo, id_operador, fecha, turno, horometro_inicial, horometro_final) VALUES
 (:'pr1', :'emp', :'m_cien', :'obra_reb', :'eq_scoop', :'e_maria','2026-06-01','M',1200,1207.5);
INSERT INTO produccion.rezagado_ciclo (id_empresa, id_parte_rezagado, desde, hasta, id_tipo_mineral, toneladas) VALUES
 (:'emp', :'pr1','Tope 2400W','Cruce',(SELECT id FROM catalogos.tipo_mineral WHERE id_empresa=:'emp' AND codigo='MINERAL'),150);

-- ---- Producción: BARRENACIÓN (larga) ----
INSERT INTO produccion.parte_barrenacion (id, id_empresa, tipo_barrenacion, id_mina, id_obra, id_equipo, id_capitan_mina, id_supervisor, id_operador, id_actividad, fecha, turno, horometro_percusion_inicial, horometro_percusion_final) VALUES
 (:'pb1', :'emp','LARGA', :'m_cien', :'obra_reb', :'eq_jum_lg', :'e_juan', :'e_cesar', :'e_pedro',(SELECT id FROM catalogos.actividad WHERE id_empresa=:'emp' AND codigo='BARRENACION'),'2026-06-01','M',2100,2108.5);
INSERT INTO produccion.barrenacion_avance (id_empresa, id_parte_barrenacion, actividad, lugar, no_secciones, longitud, no_barrenos) VALUES
 (:'emp', :'pb1','BARRENACION','Tajo 2400W Bl-5',1,10,5),
 (:'emp', :'pb1','BARRENACION','Tajo 2400W Bl-6',1,10,5);
-- "Ejecutado": secciones de barreno realizadas en el avance Bl-5 (lista fija múltiple)
INSERT INTO produccion.barrenacion_ejecutado (id_empresa, id_barrenacion_avance, id_tipo_barreno)
SELECT :'emp',
       (SELECT id FROM produccion.barrenacion_avance WHERE id_parte_barrenacion=:'pb1' AND lugar='Tajo 2400W Bl-5'),
       tb.id
FROM catalogos.tipo_barreno tb
WHERE tb.id_empresa=:'emp' AND tb.codigo IN ('CUELE','DESTROZA','CONTORNO','ZAPATERA');

-- ---- Demoras de equipo ----
INSERT INTO produccion.demora_equipo (id_empresa, id_mina, id_equipo, id_tipo_demora, fecha, turno, minutos, observacion) VALUES
 (:'emp', :'m_cien', :'eq_camion', (SELECT id FROM catalogos.tipo_demora WHERE id_empresa=:'emp' AND codigo='MECANICA'),'2026-06-01','M',60,'Cambio de llanta'),
 (:'emp', :'m_cien', :'eq_jum_lg', (SELECT id FROM catalogos.tipo_demora WHERE id_empresa=:'emp' AND codigo='ELECTRICA'),'2026-06-01','M',30,'Falla brazo');

-- ---- Reconciliación ----
INSERT INTO reconciliacion.segmento (id, id_empresa, id_mina, id_rebaje, codigo, descripcion) VALUES
 (:'seg_001', :'emp', :'m_cien', :'reb_tase','SEG-2400W-001','Segmento 10x2 tajo 2400W');
INSERT INTO reconciliacion.medicion (id_empresa, id_segmento, fuente, fecha, toneladas, ley_au, ley_ag) VALUES
 (:'emp', :'seg_001','RESERVA','2026-05-01',1000,0.85,150),
 (:'emp', :'seg_001','TUMBE','2026-06-01',950,0.80,145),
 (:'emp', :'seg_001','ESTIMACION','2026-06-05',940,0.79,143),
 (:'emp', :'seg_001','PLANTA','2026-06-10',900,0.78,140);

-- ---- Costos ----
INSERT INTO costos.contratista (id, id_empresa, codigo, nombre) VALUES (:'contra1', :'emp','C001','Contratista Desarrollo SA');
INSERT INTO costos.estimacion (id, id_empresa, id_mina, id_contratista, numero, tipo, periodo, fecha, monto_total, estado) VALUES
 (:'est1', :'emp', :'m_cien', :'contra1','EST-2026-06-001','DESARROLLO','2026-06','2026-06-15',45000,'APROBADA');
INSERT INTO costos.estimacion_concepto (id_empresa, id_estimacion, concepto, rubro, actividad, cantidad, precio_unitario, importe) VALUES
 (:'emp', :'est1','Avance lineal RPA','DESARROLLO','Barrenacion',150,200,30000),
 (:'emp', :'est1','Rezagado','DESARROLLO','Rezagado',300,50,15000);
INSERT INTO costos.presupuesto (id_empresa, id_mina, anio, periodo, rubro, monto_presupuestado, moneda) VALUES
 (:'emp', :'m_cien',2026,'2026-06','DESARROLLO',50000,'USD');
INSERT INTO costos.costo_unitario (id_empresa, id_mina, periodo, actividad, area, unidad, valor) VALUES
 (:'emp', :'m_cien','2026-06','Acarreo','Cienega','USD_TON',2.35),
 (:'emp', :'m_cien','2026-06','Desarrollo','Cienega','USD_METRO',310.50);
INSERT INTO costos.cutoff_variable (id_empresa, id_mina, id_rebaje, zona, periodo, cutoff_valor, unidad) VALUES
 (:'emp', :'m_cien', :'reb_tase','Cienega Oeste','2026-06',55.00,'USD_TON');

-- ---- Explosivo ----
INSERT INTO produccion.consumo_explosivo (id_empresa, id_mina, id_obra, id_tipo_explosivo, id_parte_barrenacion, fecha, turno, destino, cantidad) VALUES
 (:'emp', :'m_cien', :'obra_reb', (SELECT id FROM catalogos.tipo_explosivo WHERE id_empresa=:'emp' AND codigo='ANFO'), :'pb1','2026-06-01','M','TUMBE',250),
 (:'emp', :'m_cien', :'obra_reb', (SELECT id FROM catalogos.tipo_explosivo WHERE id_empresa=:'emp' AND codigo='FULMINANTE'), :'pb1','2026-06-01','M','TUMBE',10);

-- ---- Beneficio (planta de molienda) ----
INSERT INTO beneficio.lote_molienda (id, id_empresa, id_mina, codigo_lote, fecha, toneladas_molidas, estado) VALUES
 (:'lote1', :'emp', :'m_cien','LOTE-2026-06-01','2026-06-03',720,'CERRADO');
INSERT INTO beneficio.ley_metalurgica (id_empresa, id_lote_molienda, punto, ley_au, ley_ag) VALUES
 (:'emp', :'lote1','CABEZA',0.78,140),
 (:'emp', :'lote1','CONCENTRADO',45.2,7800),
 (:'emp', :'lote1','COLA',0.05,8);
INSERT INTO beneficio.recuperacion (id_empresa, id_lote_molienda, metal, recuperacion_pct) VALUES
 (:'emp', :'lote1','AU',92.5),
 (:'emp', :'lote1','AG',90.1);

-- ---- Estándares (tiempos y productividad) ----
INSERT INTO estandares.estandar_tiempo (id_empresa, id_tipo_equipo, disponibilidad_meta_pct, utilizacion_meta_pct, demora_max_pct, vigente_desde) VALUES
 (:'emp', (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='CAMION'),85,80,15,'2026-01-01'),
 (:'emp', (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='JUMBO_LARGA'),80,75,20,'2026-01-01');
INSERT INTO estandares.estandar_productividad (id_empresa, id_actividad, id_tipo_equipo, unidad, valor_meta, vigente_desde) VALUES
 (:'emp', (SELECT id FROM catalogos.actividad WHERE id_empresa=:'emp' AND codigo='BARRENACION'), (SELECT id FROM catalogos.tipo_equipo WHERE id_empresa=:'emp' AND codigo='JUMBO_LARGA'),'METRO_HORA',2.5,'2026-01-01');

-- ---- Inversiones ----
INSERT INTO inversiones.activo (id_empresa, id_mina, codigo, descripcion, tipo_activo, costo_adquisicion, fecha_adquisicion, vida_util_meses) VALUES
 (:'emp', :'m_cien','ACT-001','Jumbo Epiroc L2C','Equipo de barrenacion',850000,'2021-03-15',120);
INSERT INTO inversiones.inversion (id_empresa, id_mina, id_obra, periodo, concepto, monto) VALUES
 (:'emp', :'m_cien', :'obra_reb','2026-06','Desarrollo rebaje 2400W',120000),
 (:'emp', :'m_cien', :'obra_rpa','2026-06','Avance rampa 2140',80000);
INSERT INTO inversiones.consumo_acero (id_empresa, id_mina, id_obra, fecha, tipo_acero, cantidad, unidad) VALUES
 (:'emp', :'m_cien', :'obra_reb','2026-06-01','Split set',120,'pza'),
 (:'emp', :'m_cien', :'obra_reb','2026-06-01','Malla electrosoldada',45,'m2');

-- Estadísticas para el planificador (tras cargar datos)
ANALYZE;
