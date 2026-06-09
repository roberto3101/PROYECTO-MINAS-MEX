-- ============================================================
-- 12 · Integridad multi-tenant — FKs compuestas (id_empresa, id)
-- ============================================================
-- El aislamiento de tenants se garantiza EN LA BASE, no en la app:
--   (1) Cada tabla padre con tenant expone UNIQUE (id_empresa, id).
--   (2) Toda FK de negocio entre tablas con id_empresa se convierte a
--       FK compuesta: (id_empresa, id_x) -> padre (id_empresa, id).
-- Resultado: una fila NUNCA puede referenciar datos de otra empresa,
-- aunque el código de la aplicación tenga errores.
-- Notas: las columnas de auditoría *_por_usuario_id no llevan FK por
-- diseño; la FK id_empresa -> gobierno.empresa(id) se mantiene simple
-- (es el ancla del tenant). Idempotente: solo convierte FKs de 1 columna.

DO $$
DECLARE
  p RECORD;
  f RECORD;
  v_uq  int := 0;
  v_fk  int := 0;
BEGIN
  -- (1) UNIQUE (id_empresa, id) en cada padre de negocio referenciado
  FOR p IN
    SELECT DISTINCT pn.nspname AS sch, pc.relname AS tbl, c.confrelid AS reloid
    FROM pg_constraint c
    JOIN pg_class pc      ON pc.oid = c.confrelid
    JOIN pg_namespace pn  ON pn.oid = pc.relnamespace
    JOIN pg_class cc      ON cc.oid = c.conrelid
    JOIN pg_namespace cn  ON cn.oid = cc.relnamespace
    WHERE c.contype = 'f'
      AND cn.nspname IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
      AND pn.nspname IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
      AND EXISTS (SELECT 1 FROM pg_attribute a
                  WHERE a.attrelid = c.confrelid AND a.attname = 'id_empresa'
                    AND a.attnum > 0 AND NOT a.attisdropped)
  LOOP
    IF NOT EXISTS (SELECT 1 FROM pg_constraint u
                   WHERE u.conrelid = p.reloid AND u.conname = 'uq_'||p.tbl||'_tenant') THEN
      EXECUTE format('ALTER TABLE %I.%I ADD CONSTRAINT %I UNIQUE (id_empresa, id)',
                     p.sch, p.tbl, 'uq_'||p.tbl||'_tenant');
      v_uq := v_uq + 1;
    END IF;
  END LOOP;

  -- (2) Convertir cada FK simple de negocio en FK compuesta con el tenant
  FOR f IN
    SELECT c.conname,
           cn.nspname AS csch, cc.relname AS ctbl,
           pn.nspname AS psch, pc.relname AS ptbl,
           a.attname  AS col
    FROM pg_constraint c
    JOIN pg_class cc      ON cc.oid = c.conrelid
    JOIN pg_namespace cn  ON cn.oid = cc.relnamespace
    JOIN pg_class pc      ON pc.oid = c.confrelid
    JOIN pg_namespace pn  ON pn.oid = pc.relnamespace
    JOIN pg_attribute a   ON a.attrelid = c.conrelid AND a.attnum = c.conkey[1]
    WHERE c.contype = 'f'
      AND array_length(c.conkey, 1) = 1
      AND a.attname <> 'id_empresa'
      AND cn.nspname IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
      AND pn.nspname IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
      AND EXISTS (SELECT 1 FROM pg_attribute x
                  WHERE x.attrelid = c.conrelid AND x.attname = 'id_empresa'
                    AND x.attnum > 0 AND NOT x.attisdropped)
  LOOP
    EXECUTE format('ALTER TABLE %I.%I DROP CONSTRAINT %I', f.csch, f.ctbl, f.conname);
    EXECUTE format('ALTER TABLE %I.%I ADD CONSTRAINT %I FOREIGN KEY (id_empresa, %I) REFERENCES %I.%I (id_empresa, id) ON DELETE RESTRICT',
                   f.csch, f.ctbl, f.conname, f.col, f.psch, f.ptbl);
    v_fk := v_fk + 1;
  END LOOP;

  RAISE NOTICE 'Blindaje tenant: % UNIQUE(id_empresa,id) creados, % FKs convertidas a compuestas', v_uq, v_fk;
END $$;

-- FKs compuestas explícitas de gobierno hacia catálogos (los padres ya
-- tienen su UNIQUE (id_empresa, id) por el bloque dinámico de arriba):
ALTER TABLE gobierno.usuario
  ADD CONSTRAINT fk_usuario_empleado FOREIGN KEY (id_empresa, id_empleado)
  REFERENCES catalogos.empleado (id_empresa, id) ON DELETE RESTRICT;
ALTER TABLE gobierno.usuario_rol
  ADD CONSTRAINT fk_usuario_rol_mina FOREIGN KEY (id_empresa, id_mina)
  REFERENCES catalogos.mina (id_empresa, id) ON DELETE RESTRICT;

-- Autoverificación: no debe quedar NINGUNA FK simple de negocio entre tablas tenant
DO $$
DECLARE v_rest int;
BEGIN
  SELECT count(*) INTO v_rest
  FROM pg_constraint c
  JOIN pg_class cc      ON cc.oid = c.conrelid
  JOIN pg_namespace cn  ON cn.oid = cc.relnamespace
  JOIN pg_class pc      ON pc.oid = c.confrelid
  JOIN pg_namespace pn  ON pn.oid = pc.relnamespace
  JOIN pg_attribute a   ON a.attrelid = c.conrelid AND a.attnum = c.conkey[1]
  WHERE c.contype = 'f' AND array_length(c.conkey,1) = 1
    AND a.attname <> 'id_empresa'
    AND cn.nspname IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones')
    AND pn.nspname IN ('catalogos','produccion','planeacion','reconciliacion','beneficio','estandares','costos','inversiones');
  IF v_rest <> 0 THEN
    RAISE EXCEPTION 'Blindaje tenant INCOMPLETO: quedan % FKs simples de negocio', v_rest;
  END IF;
  RAISE NOTICE 'Verificado: 0 FKs simples de negocio restantes (todas compuestas con id_empresa)';
END $$;
