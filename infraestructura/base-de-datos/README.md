# Base de datos — Sistema de Planeación y Producción Minera

Esquema **PostgreSQL 17**, organizado por **capacidades** (schemas), con DDD ligero:
UUID, multi-tenant, auditoría inline, eventos append-only, índices únicos parciales y
vistas de reportes. El **aislamiento multi-tenant vive en la base** (RLS + FK compuestas
+ privilegios), no en el código de la aplicación. Todo el pipeline (esquema → vistas →
seed → tests) se ejecuta y **pasa en local** (29 pruebas, auditada además por subagentes senior).

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
├── 02_catalogos.sql          (16) mina, equipo, empleado, mineral, obra + 11 tipificaciones
├── 03_produccion.sql         (9)  parte_acarreo+acarreo_viaje, parte_rezagado+rezagado_ciclo,
│                                   parte_barrenacion+barrenacion_avance+barrenacion_ejecutado,
│                                   demora_equipo, consumo_explosivo*
├── 04_planeacion.sql         (4)  rebaje, plan, bloque_programado, meta_periodo
├── 05_reconciliacion.sql     (2)  segmento (10x2), medicion (reserva/tumbe/estimación/planta)
├── 06_costos.sql             (6)  contratista, estimacion, estimacion_concepto, presupuesto,
│                                   costo_unitario, cutoff_variable
├── 07_explosivo.sql          (*)  tipo_explosivo + consumo_explosivo (control de explosivo)
├── 08_beneficio.sql          (3)  lote_molienda, ley_metalurgica (cabeza/conc/cola), recuperacion
├── 09_estandares.sql         (2)  estandar_tiempo, estandar_productividad
├── 10_inversiones.sql        (3)  activo, inversion, consumo_acero
├── 11_indices.sql            índices de FK + parciales para escala (rendimiento)
├── 12_integridad_tenant.sql  FKs COMPUESTAS (id_empresa, id): imposible referenciar otro tenant
├── 50_vistas.sql             19 vistas + 1 materializada (capa reportes)
└── 60_seguridad_rls.sql      rol `aplicacion`, RLS fail-closed, security_invoker, privilegios
semilla/20_seed.sql           datos de ejemplo (Excel real + tenant B mínimo para aislamiento)
pruebas/30_tests.sql          29 tests: integración, lógica, gobernanza y aislamiento multi-tenant
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
- **Aislamiento multi-tenant EN la base** (no en la app): ver sección siguiente.

## Seguridad multi-tenant (en la base)
El backend **no** escribe `WHERE id_empresa` en ninguna consulta. Tres capas lo garantizan:

1. **RLS fail-closed** (test T24-T26): política `p_tenant` en toda tabla con `id_empresa`.
   Sin tenant fijado en la sesión → 0 filas y todo INSERT rechazado. Las vistas la respetan
   (`security_invoker = on`) y la materializada solo se expone vía `reportes.v_plan_vs_real_mensual`.
2. **FK compuestas** `(id_empresa, id)` (test T27): ninguna fila puede referenciar datos de
   otra empresa — ni siquiera conexiones administrativas que esquivan RLS.
3. **Privilegios del rol `aplicacion`** (test T28): sin `BYPASSRLS`, sin `DELETE` (borrado
   lógico por `UPDATE eliminado_en`), y los eventos sin `UPDATE` (append-only por privilegio).

Patrón del backend (por transacción):
```sql
BEGIN;
SELECT set_config('app.empresa_actual', $1, true);  -- uuid del tenant (del JWT); true = local a la tx
SET LOCAL ROLE aplicacion;
-- ... consultas normales, sin WHERE id_empresa ...
COMMIT;  -- el tenant y el rol se limpian al cerrar la transacción (seguro con pool de conexiones)
```
El rol LOGIN real del backend se crea como miembro: `CREATE ROLE backend LOGIN PASSWORD '...' IN ROLE aplicacion;`

## Cobertura
Cubre y prueba las áreas del PDF: catálogos, captura del Real, plan de bloques (6 horizontes),
reconciliación, planta de beneficio, estándares de tiempos/productividad, costos/presupuesto/
cut-off, explosivo e inversiones. Los ~2,000 informes son parametrizaciones de las vistas base.
Pendiente (no es base de datos): capa de aplicación Go, frontend y motor de sync.
