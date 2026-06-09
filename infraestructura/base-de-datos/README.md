# Base de datos — Sistema de Planeación y Producción Minera

Esquema **PostgreSQL 17**, organizado por **capacidades** (schemas), con DDD ligero:
UUID, multi-tenant, auditoría inline, eventos append-only, índices únicos parciales y
vistas de reportes. Todo el pipeline (esquema → vistas → seed → tests) se ejecuta y
**pasa en local** (22 tests, auditada además por subagentes senior).

## Cómo ejecutar

### 1. Cluster de pruebas aislado (no toca tu PostgreSQL del sistema)
```powershell
$PG='C:\Program Files\PostgreSQL\17\bin'
$DATA='C:\Users\user\Desktop\MinasSali\.devdb\data'
& "$PG\initdb.exe" -D $DATA -U postgres -A trust -E UTF8 --locale=C
& "$PG\pg_ctl.exe" -D $DATA -l "$DATA\..\log.txt" -o "-p 5433" -w start
& "$PG\psql.exe" -U postgres -h 127.0.0.1 -p 5433 -d postgres -c "CREATE DATABASE mina;"
```

### 2. Cargar esquema + seed + tests
```powershell
cd C:\Users\user\Desktop\MinasSali\infraestructura\base-de-datos
.\ejecutar.ps1                       # cluster de pruebas (5433/mina)
.\ejecutar.ps1 -SoloEsquema          # solo DDL + vistas (sin datos, para producción)
```
Imprime `TODOS LOS TESTS PASARON` al final si todo está correcto.

## Páginas (capacidades) — 47 tablas
```
infraestructura/base-de-datos/esquema/
├── 00_reset_y_esquemas.sql   schemas por capacidad (idempotente)
├── 01_gobierno.sql           (2)  empresa, usuario
├── 02_catalogos.sql          (15) mina, equipo, empleado, mineral, obra + 10 tipificaciones
├── 03_produccion.sql         (8)  parte_acarreo+acarreo_viaje, parte_rezagado+rezagado_ciclo,
│                                   parte_barrenacion+barrenacion_avance, demora_equipo, consumo_explosivo*
├── 04_planeacion.sql         (4)  rebaje, plan, bloque_programado, meta_periodo
├── 05_reconciliacion.sql     (2)  segmento (10x2), medicion (reserva/tumbe/estimación/planta)
├── 06_costos.sql             (6)  contratista, estimacion, estimacion_concepto, presupuesto,
│                                   costo_unitario, cutoff_variable
├── 07_explosivo.sql          (*)  tipo_explosivo + consumo_explosivo (control de explosivo)
├── 08_beneficio.sql          (3)  lote_molienda, ley_metalurgica (cabeza/conc/cola), recuperacion
├── 09_estandares.sql         (2)  estandar_tiempo, estandar_productividad
├── 10_inversiones.sql        (3)  activo, inversion, consumo_acero
├── 11_indices.sql            índices de FK + parciales para escala (rendimiento)
└── 50_vistas.sql             18 vistas + 1 materializada (capa reportes)
semilla/20_seed.sql           datos de ejemplo (Excel real + plan/molienda/costos/inversiones)
pruebas/30_tests.sql          22 tests de integración, lógica y gobernanza
pruebas/agentes/              baterías de auditoría senior (smoke · integridad · lógica · rendimiento · trazabilidad · dominio)
ejecutar.ps1 · README.md
```
\* `consumo_explosivo` vive en el schema `produccion` (su catálogo `tipo_explosivo` en `catalogos`).

## Lenguaje ubicuo (nombres específicos del dominio)
`acarreo_viaje` (viaje de camión), `rezagado_ciclo` (ciclo de scooptram), `barrenacion_avance`
(avance barrenado), `demora_equipo`, `obra` (labor minera, entidad central), `bloque_programado`,
`meta_periodo`, `lote_molienda`, `ley_metalurgica` (cabeza/concentrado/cola), `cutoff_variable`.

## Vistas de reportes (las familias de los "+2,000 entregables")
producción acarreo/rezagado, balance de cargas, avance de barrenación, horas/disponibilidad de
equipo, **tiempos operativos vs estándar**, productividad, plan por periodo, **plan vs real**
(materializada), obras prioritarias, reconciliación, **producción vs molienda**, balance
metalúrgico, costo de estimación, presupuesto vs real, costo unitario, consumo de explosivo,
inversión por obra. Cada "informe" del PDF es una de estas filtrada por periodo/mina/obra.

## Decisiones de diseño
- **UUID** PK; **multi-tenant** `id_empresa` en toda tabla (test T12).
- **Append-only** en eventos (viaje/ciclo/avance/medición/consumo) — solo `creado_*` (test T13).
- **Auditoría inline**: `creado/actualizado/eliminado_en` + `_por_usuario_id`. Sin FK a `usuario`
  (el actor puede ser el proceso de sync). Borrado **lógico** (`eliminado_en IS NULL` = activo).
- **Índices únicos parciales** `WHERE eliminado_en IS NULL` → único entre activos, permite re-alta
  tras baja lógica (test T15).
- **Eventos se compensan, no se borran**; borrado físico solo como purga administrativa.
- **Discriminadores**: `tipo_barrenacion` (4 tablas del ER → 1), `actividad`.
- Lógica de negocio en **Go** (no procedimientos almacenados).

## Cobertura
Cubre y prueba las áreas del PDF: catálogos, captura del Real, plan de bloques (6 horizontes),
reconciliación, planta de beneficio, estándares de tiempos/productividad, costos/presupuesto/
cut-off, explosivo e inversiones. Los ~2,000 informes son parametrizaciones de las vistas base.
Pendiente (no es base de datos): capa de aplicación Go, frontend y motor de sync.
