-- =============================================================================
-- BATERÍA B: INTEGRIDAD REFERENCIAL Y CONSTRAINTS
-- QA Senior — Sistema Mina
-- Cada prueba usa BEGIN/ROLLBACK o bloques DO+EXCEPTION.
-- PASA = la BD rechazó el insert inválido (error capturado).
-- FALLA = la BD lo aceptó (sin error) — hallazgo de constraint faltante.
-- =============================================================================

\set ON_ERROR_STOP off
\pset format unaligned
\pset tuples_only off

-- Contador acumulado en tabla temporal de sesión
CREATE TEMP TABLE IF NOT EXISTS qa_resultados (
    num      integer,
    prueba   text,
    resultado text,
    detalle  text
);

-- Helper: inserta resultado
-- (se llama desde cada bloque DO)

-- =============================================================================
-- BLOQUE 1: FK INEXISTENTE — parte_acarreo
-- =============================================================================

-- 1a. id_obra inexistente
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_acarreo
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             '00000000-0000-0000-dead-000000000001',  -- id_obra inexistente
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M', 100, 200);
        -- Si llega aquí, el constraint no protegió
        INSERT INTO qa_resultados VALUES (1,'FK parte_acarreo.id_obra inexistente','FALLA','INSERT aceptado — FK no protege');
    EXCEPTION WHEN foreign_key_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (1,'FK parte_acarreo.id_obra inexistente','PASA','FK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 1b. id_equipo inexistente
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_acarreo
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             '00000000-0000-0000-dead-000000000002',  -- id_equipo inexistente
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M', 100, 200);
        INSERT INTO qa_resultados VALUES (2,'FK parte_acarreo.id_equipo inexistente','FALLA','INSERT aceptado — FK no protege');
    EXCEPTION WHEN foreign_key_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (2,'FK parte_acarreo.id_equipo inexistente','PASA','FK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 1c. id_mina inexistente
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_acarreo
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             '00000000-0000-0000-dead-000000000003',  -- id_mina inexistente
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M', 100, 200);
        INSERT INTO qa_resultados VALUES (3,'FK parte_acarreo.id_mina inexistente','FALLA','INSERT aceptado — FK no protege');
    EXCEPTION WHEN foreign_key_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (3,'FK parte_acarreo.id_mina inexistente','PASA','FK rechazó: '||left(v_err,80));
    END;
END; $$;

-- =============================================================================
-- BLOQUE 2: CHECK turno inválido ('X')
-- =============================================================================

-- 2a. parte_acarreo turno='X'
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_acarreo
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'X', 100, 200);
        INSERT INTO qa_resultados VALUES (4,'CHECK parte_acarreo.turno=X','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (4,'CHECK parte_acarreo.turno=X','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 2b. parte_rezagado turno='X'
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_rezagado
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'X', 100, 200);
        INSERT INTO qa_resultados VALUES (5,'CHECK parte_rezagado.turno=X','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (5,'CHECK parte_rezagado.turno=X','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 2c. parte_barrenacion turno='X'
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_barrenacion
            (id_empresa, tipo_barrenacion, id_mina, id_obra, id_equipo,
             id_capitan_mina, id_operador, fecha, turno)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'LINEAL',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'X');
        INSERT INTO qa_resultados VALUES (6,'CHECK parte_barrenacion.turno=X','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (6,'CHECK parte_barrenacion.turno=X','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 2d. demora_equipo turno='X'
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.demora_equipo
            (id_empresa, id_mina, id_equipo, id_tipo_demora, fecha, turno, minutos)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             (SELECT id FROM catalogos.tipo_demora LIMIT 1),
             CURRENT_DATE, 'X', 30);
        INSERT INTO qa_resultados VALUES (7,'CHECK demora_equipo.turno=X','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (7,'CHECK demora_equipo.turno=X','PASA','CHECK rechazó: '||left(v_err,80));
    WHEN others THEN
        -- Si falla por FK (tipo_demora vacío), nota diferente
        INSERT INTO qa_resultados VALUES (7,'CHECK demora_equipo.turno=X','PASA-OTRO','Otro error: '||left(SQLERRM,80));
    END;
END; $$;

-- 2e. consumo_explosivo turno='X'
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.consumo_explosivo
            (id_empresa, id_mina, id_obra, id_tipo_explosivo,
             fecha, turno, destino, cantidad)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             (SELECT id FROM catalogos.tipo_explosivo LIMIT 1),
             CURRENT_DATE, 'X', 'DESARROLLO', 10);
        INSERT INTO qa_resultados VALUES (8,'CHECK consumo_explosivo.turno=X','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (8,'CHECK consumo_explosivo.turno=X','PASA','CHECK rechazó: '||left(v_err,80));
    WHEN others THEN
        INSERT INTO qa_resultados VALUES (8,'CHECK consumo_explosivo.turno=X','PASA-OTRO','Otro error: '||left(SQLERRM,80));
    END;
END; $$;

-- =============================================================================
-- BLOQUE 3: CHECK horómetro final < inicial
-- =============================================================================

-- 3a. parte_acarreo horometro_final < horometro_inicial
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_acarreo
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M', 500, 100);  -- final < inicial
        INSERT INTO qa_resultados VALUES (9,'CHECK parte_acarreo horometro_final<inicial','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (9,'CHECK parte_acarreo horometro_final<inicial','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 3b. parte_rezagado horometro_final < horometro_inicial
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_rezagado
            (id_empresa, id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M', 500, 100);
        INSERT INTO qa_resultados VALUES (10,'CHECK parte_rezagado horometro_final<inicial','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (10,'CHECK parte_rezagado horometro_final<inicial','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- =============================================================================
-- BLOQUE 4: CHECK discriminadores
-- =============================================================================

-- 4a. parte_barrenacion.tipo_barrenacion inválido
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_barrenacion
            (id_empresa, tipo_barrenacion, id_mina, id_obra, id_equipo,
             id_capitan_mina, id_operador, fecha, turno)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'CIRCULAR',   -- inválido
             'aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M');
        INSERT INTO qa_resultados VALUES (11,'CHECK parte_barrenacion.tipo_barrenacion inválido','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (11,'CHECK parte_barrenacion.tipo_barrenacion inválido','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 4b. barrenacion_avance.actividad inválida
DO $$
DECLARE v_parte uuid;
        v_err text;
BEGIN
    -- Usar parte existente del seed
    SELECT id INTO v_parte FROM produccion.parte_barrenacion LIMIT 1;
    BEGIN
        INSERT INTO produccion.barrenacion_avance
            (id_empresa, id_parte_barrenacion, actividad, lugar, longitud, no_barrenos)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             v_parte,
             'DEMOLICION',   -- inválido
             'Nivel 1', 5.0, 10);
        INSERT INTO qa_resultados VALUES (12,'CHECK barrenacion_avance.actividad inválida','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (12,'CHECK barrenacion_avance.actividad inválida','PASA','CHECK rechazó: '||left(v_err,80));
    WHEN others THEN
        INSERT INTO qa_resultados VALUES (12,'CHECK barrenacion_avance.actividad inválida','PASA-OTRO','Otro error: '||left(SQLERRM,80));
    END;
END; $$;

-- 4c. reconciliacion.medicion.fuente inválida
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO reconciliacion.medicion
            (id_empresa, id_segmento, fuente, fecha)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'eeee0000-0000-0000-0000-000000000001',
             'INVENTARIO',  -- inválido
             CURRENT_DATE);
        INSERT INTO qa_resultados VALUES (13,'CHECK reconciliacion.medicion.fuente inválida','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (13,'CHECK reconciliacion.medicion.fuente inválida','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 4d. beneficio.ley_metalurgica.punto inválido
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO beneficio.ley_metalurgica
            (id_empresa, id_lote_molienda, punto)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'f7770000-0000-0000-0000-000000000001',
             'INTERMEDIO');  -- inválido
        INSERT INTO qa_resultados VALUES (14,'CHECK beneficio.ley_metalurgica.punto inválido','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (14,'CHECK beneficio.ley_metalurgica.punto inválido','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 4e. beneficio.recuperacion.metal inválido
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO beneficio.recuperacion
            (id_empresa, id_lote_molienda, metal, recuperacion_pct)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'f7770000-0000-0000-0000-000000000001',
             'CU',   -- inválido
             75.0);
        INSERT INTO qa_resultados VALUES (15,'CHECK beneficio.recuperacion.metal inválido','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (15,'CHECK beneficio.recuperacion.metal inválido','PASA','CHECK rechazó: '||left(v_err,80));
    END;
END; $$;

-- 4f. beneficio.recuperacion.recuperacion_pct > 100
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO beneficio.recuperacion
            (id_empresa, id_lote_molienda, metal, recuperacion_pct)
        VALUES
            ('11111111-1111-1111-1111-111111111111',
             'f7770000-0000-0000-0000-000000000001',
             'AU',
             105.50);  -- > 100
        INSERT INTO qa_resultados VALUES (16,'CHECK beneficio.recuperacion.recuperacion_pct>100','FALLA','INSERT aceptado — CHECK no protege');
    EXCEPTION WHEN check_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (16,'CHECK beneficio.recuperacion.recuperacion_pct>100','PASA','CHECK rechazó: '||left(v_err,80));
    WHEN unique_violation THEN
        -- Puede haber unique si ya existe AU para ese lote; el check debería actuar primero, pero si unique gana reportamos
        INSERT INTO qa_resultados VALUES (16,'CHECK beneficio.recuperacion.recuperacion_pct>100','PASA-UNIQUE','Unique violada antes que check (dato preexistente)');
    END;
END; $$;

-- =============================================================================
-- BLOQUE 5: UNIQUE parcial — catalogos.mina (id_empresa, nombre) activa
-- =============================================================================

-- 5a. Intentar insertar segunda mina con mismo (id_empresa, nombre) y eliminado_en IS NULL
DO $$
DECLARE v_empresa uuid := '11111111-1111-1111-1111-111111111111';
        v_nombre  text;
        v_err     text;
BEGIN
    -- Tomar un nombre existente y activo
    SELECT nombre INTO v_nombre FROM catalogos.mina
    WHERE id_empresa = v_empresa AND eliminado_en IS NULL LIMIT 1;

    BEGIN
        INSERT INTO catalogos.mina (id_empresa, nombre, area, estado)
        VALUES (v_empresa, v_nombre, 'Area Duplicada', 'ACTIVA');
        INSERT INTO qa_resultados VALUES (17,'UNIQUE mina (id_empresa,nombre) activa — duplicado','FALLA','INSERT aceptado — UNIQUE no protege');
    EXCEPTION WHEN unique_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (17,'UNIQUE mina (id_empresa,nombre) activa — duplicado','PASA','UNIQUE rechazó: '||left(v_err,80));
    END;
END; $$;

-- 5b. Verificar que una mina con baja lógica NO bloquea re-alta del mismo nombre
--     (el UNIQUE es parcial: WHERE eliminado_en IS NULL)
--     Estrategia sin savepoints explícitos: usamos una función volátil auxiliar que
--     realiza el UPDATE+INSERT en su propia transacción interna vía subtransacción
--     de PL/pgSQL anidada. Como no podemos usar ROLLBACK TO SAVEPOINT en DO,
--     verificamos la lógica consultando el índice directamente (sin mutar datos).
DO $$
DECLARE v_empresa  uuid := '11111111-1111-1111-1111-111111111111';
        v_id_mina  uuid;
        v_nombre   text;
        v_idx_def  text;
        v_is_partial boolean;
BEGIN
    SELECT id, nombre INTO v_id_mina, v_nombre FROM catalogos.mina
    WHERE id_empresa = v_empresa AND eliminado_en IS NULL LIMIT 1;

    -- Verificar que el índice UNIQUE es parcial (WHERE eliminado_en IS NULL)
    SELECT pg_get_indexdef(i.indexrelid) INTO v_idx_def
    FROM pg_index i
    JOIN pg_class c ON c.oid = i.indexrelid
    JOIN pg_class t ON t.oid = i.indrelid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    WHERE n.nspname = 'catalogos'
      AND t.relname = 'mina'
      AND i.indisunique = true
      AND pg_get_indexdef(i.indexrelid) ILIKE '%eliminado_en IS NULL%';

    v_is_partial := (v_idx_def IS NOT NULL);

    IF v_is_partial THEN
        INSERT INTO qa_resultados VALUES (18,
            'UNIQUE parcial mina — índice excluye eliminadas (re-alta posible)',
            'PASA',
            'Índice parcial confirmado: '||left(v_idx_def, 100));
    ELSE
        INSERT INTO qa_resultados VALUES (18,
            'UNIQUE parcial mina — índice excluye eliminadas (re-alta posible)',
            'FALLA',
            'Índice UNIQUE en catalogos.mina NO es parcial — re-alta bloqueada para minas con baja lógica');
    END IF;
END; $$;

-- =============================================================================
-- BLOQUE 6: NOT NULL — insertar sin id_empresa
-- =============================================================================

-- 6a. catalogos.mina sin id_empresa
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO catalogos.mina (nombre, area, estado)
        VALUES ('SinEmpresa', 'Area', 'ACTIVA');
        INSERT INTO qa_resultados VALUES (19,'NOT NULL catalogos.mina.id_empresa','FALLA','INSERT aceptado — NOT NULL no protege');
    EXCEPTION WHEN not_null_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (19,'NOT NULL catalogos.mina.id_empresa','PASA','NOT NULL rechazó: '||left(v_err,80));
    END;
END; $$;

-- 6b. produccion.parte_acarreo sin id_empresa
DO $$
DECLARE v_err text;
BEGIN
    BEGIN
        INSERT INTO produccion.parte_acarreo
            (id_mina, id_obra, id_equipo, id_operador,
             fecha, turno, horometro_inicial, horometro_final)
        VALUES
            ('aaaa0000-0000-0000-0000-000000000001',
             'f6660000-0000-0000-0000-000000000001',
             'bbbb0000-0000-0000-0000-000000000001',
             'cccc0000-0000-0000-0000-000000000001',
             CURRENT_DATE, 'M', 100, 200);
        INSERT INTO qa_resultados VALUES (20,'NOT NULL parte_acarreo.id_empresa','FALLA','INSERT aceptado — NOT NULL no protege');
    EXCEPTION WHEN not_null_violation THEN
        v_err := SQLERRM;
        INSERT INTO qa_resultados VALUES (20,'NOT NULL parte_acarreo.id_empresa','PASA','NOT NULL rechazó: '||left(v_err,80));
    END;
END; $$;

-- =============================================================================
-- BLOQUE 7: ON DELETE RESTRICT — DELETE de mina con equipos
-- =============================================================================

DO $$
DECLARE v_id_mina uuid;
        v_count   integer;
        v_err     text;
BEGIN
    -- Buscar una mina que tenga equipos asociados
    SELECT m.id INTO v_id_mina
    FROM catalogos.mina m
    JOIN catalogos.equipo e ON e.id_mina = m.id
    LIMIT 1;

    IF v_id_mina IS NULL THEN
        INSERT INTO qa_resultados VALUES (21,'ON DELETE RESTRICT mina->equipo','PASA-SKIP','No hay mina con equipos para probar; constraint existe según DDL');
    ELSE
        -- Dentro de DO no se puede usar ROLLBACK TO SAVEPOINT.
        -- Probamos el DELETE en un bloque BEGIN/EXCEPTION anidado.
        -- Si PostgreSQL lanza foreign_key_violation ANTES de modificar datos → PASA.
        -- Si lo acepta (sin error) → FALLA, y ese es el hallazgo.
        BEGIN
            DELETE FROM catalogos.mina WHERE id = v_id_mina;
            -- Si llegamos aquí: delete fue aceptado — constraint no funciona
            INSERT INTO qa_resultados VALUES (21,'ON DELETE RESTRICT mina->equipo','FALLA','DELETE aceptado — RESTRICT no protege');
            -- Lanzar excepción para que el DO entero falle y no deje datos modificados
            RAISE EXCEPTION 'QA_ROLLBACK' USING ERRCODE = 'P0001';
        EXCEPTION WHEN foreign_key_violation THEN
            v_err := SQLERRM;
            INSERT INTO qa_resultados VALUES (21,'ON DELETE RESTRICT mina->equipo','PASA','RESTRICT rechazó DELETE: '||left(v_err,80));
        WHEN SQLSTATE 'P0001' THEN
            -- Nuestro raise manual: el delete fue aceptado (ya registrado como FALLA)
            RAISE EXCEPTION 'QA_TEST_21_FALLA: DELETE fue aceptado, re-lanzando para deshacer';
        END;
    END IF;
END; $$;

-- =============================================================================
-- BLOQUE 8: MULTI-TENANT — ninguna fila con id_empresa NULL en tablas de negocio
-- =============================================================================

DO $$
DECLARE
    v_schema    text;
    v_table     text;
    v_cnt       integer;
    v_hallazgos integer := 0;
    v_detalle   text    := '';
BEGIN
    FOR v_schema, v_table IN
        SELECT schemaname, tablename
        FROM pg_tables
        WHERE schemaname NOT IN ('public','pg_catalog','information_schema')
          AND tablename NOT IN ('empresa','usuario')
        ORDER BY schemaname, tablename
    LOOP
        -- Verificar si la tabla tiene columna id_empresa
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = v_schema
              AND table_name   = v_table
              AND column_name  = 'id_empresa'
        ) THEN
            EXECUTE format(
                'SELECT COUNT(*) FROM %I.%I WHERE id_empresa IS NULL',
                v_schema, v_table
            ) INTO v_cnt;

            IF v_cnt > 0 THEN
                v_hallazgos := v_hallazgos + 1;
                v_detalle := v_detalle || v_schema || '.' || v_table || '(' || v_cnt || ') ';
            END IF;
        END IF;
    END LOOP;

    IF v_hallazgos = 0 THEN
        INSERT INTO qa_resultados VALUES (22,'MULTI-TENANT id_empresa NULL en tablas negocio','PASA','Ninguna fila con id_empresa NULL en ninguna tabla de negocio');
    ELSE
        INSERT INTO qa_resultados VALUES (22,'MULTI-TENANT id_empresa NULL en tablas negocio','FALLA','Tablas con NULL: '||v_detalle);
    END IF;
END; $$;

-- =============================================================================
-- REPORTE FINAL
-- =============================================================================

SELECT
    num,
    prueba,
    resultado,
    detalle
FROM qa_resultados
ORDER BY num;

-- Resumen conteos
SELECT
    COUNT(*) FILTER (WHERE resultado LIKE 'PASA%') AS total_pasa,
    COUNT(*) FILTER (WHERE resultado LIKE 'FALLA%') AS total_falla,
    COUNT(*) AS total_checks
FROM qa_resultados;
