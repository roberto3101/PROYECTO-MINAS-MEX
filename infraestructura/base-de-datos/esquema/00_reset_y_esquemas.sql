-- ============================================================
-- 00 · Reset e inicialización de esquemas por capacidad
-- Sistema de Planeación y Producción Minera · PostgreSQL 17
-- ============================================================
-- Idempotente para pruebas: limpia y recrea los esquemas.
-- ADVERTENCIA: el DROP borra todo el contenido de las capacidades.

CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid()

DROP SCHEMA IF EXISTS reportes       CASCADE;
DROP SCHEMA IF EXISTS inversiones    CASCADE;
DROP SCHEMA IF EXISTS estandares     CASCADE;
DROP SCHEMA IF EXISTS beneficio      CASCADE;
DROP SCHEMA IF EXISTS costos         CASCADE;
DROP SCHEMA IF EXISTS reconciliacion CASCADE;
DROP SCHEMA IF EXISTS planeacion     CASCADE;
DROP SCHEMA IF EXISTS produccion     CASCADE;
DROP SCHEMA IF EXISTS catalogos      CASCADE;
DROP SCHEMA IF EXISTS gobierno       CASCADE;

CREATE SCHEMA gobierno;        -- multi-tenant + auditoría
CREATE SCHEMA catalogos;       -- maestros: mina, equipo, empleado, mineral, obra, lookups
CREATE SCHEMA produccion;      -- captura del "Real" turno a turno (offline-first)
CREATE SCHEMA planeacion;      -- plan maestro de bloques/rebajes con metas por periodo
CREATE SCHEMA reconciliacion;  -- reservas -> tumbe -> estimación -> planta (segmentos 10x2)
CREATE SCHEMA beneficio;       -- planta de beneficio: recepción, molienda, leyes, recuperación
CREATE SCHEMA estandares;      -- metas de tiempos/movimientos y productividades
CREATE SCHEMA costos;          -- estimaciones, costos unitarios, cut-off, presupuesto
CREATE SCHEMA inversiones;     -- inversión por obra/periodo, activos, acero
CREATE SCHEMA reportes;        -- vistas/materializadas de entregables (derivados)

COMMENT ON SCHEMA produccion IS 'Captura del Real. Cabecera (parte_*) = Raiz; detalle (viaje/ciclo/avance) = Evento append-only.';
COMMENT ON SCHEMA beneficio IS 'Planta de beneficio: balance metalúrgico cabeza/concentrado/cola y recuperación.';
COMMENT ON SCHEMA estandares IS 'Metas de disponibilidad/utilización/demora y productividad por actividad.';
