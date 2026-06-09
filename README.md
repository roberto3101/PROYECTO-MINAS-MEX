# PROYECTO-MINAS-MEX

Sistema de **Planeación y Producción Minera** (mina subterránea polimetálica: Au · Ag · Pb · Zn).
Backend en **Go** (próximo) sobre **PostgreSQL**, organizado por **capacidades** con DDD ligero,
multi-tenant, eventos append-only, auditoría inline y borrado lógico.

## Estructura del repositorio

```
PROYECTO-MINAS-MEX/
├── documentacion/                 Sitio HTML del modelo de datos (se despliega en Vercel)
│   ├── index.html                 portada (7 capacidades)
│   ├── 01..07-*.html              una página por capacidad
│   ├── estilos.css · scripts.js   compartidos
│   ├── vercel.json                config de despliegue
│   └── README.md                  cómo desplegar en Vercel
└── infraestructura/
    └── base-de-datos/             Esquema PostgreSQL (fuente de verdad)
        ├── esquema/               00..60 DDL por capacidad + índices + vistas + seguridad RLS
        ├── semilla/20_seed.sql    datos de ejemplo
        ├── pruebas/30_tests.sql   29 tests: integración/lógica/gobernanza/aislamiento
        ├── pruebas/agentes/       baterías de auditoría senior
        ├── ejecutar.ps1           carga todo en orden
        └── README.md              detalle de la BD
```

> La carpeta `documentacion/` está **separada a propósito** del resto del proyecto para
> poder desplegarla sola en Vercel sin exponer el esquema ni el backend.

## Base de datos
PostgreSQL 17. **47 tablas** en 9 capacidades (schemas: gobierno, catalogos, produccion,
planeacion, reconciliacion, beneficio, estandares, costos, inversiones), **19 vistas**
(+1 materializada) y **29 tests** que pasan en local.

El **aislamiento multi-tenant vive en la base**, no en el código: RLS fail-closed por
tenant, FKs compuestas `(id_empresa, id)` que hacen imposible referenciar datos de otra
empresa, y un rol `aplicacion` sin `DELETE` (borrado lógico) ni `UPDATE` sobre eventos
(append-only por privilegio). El backend solo fija el tenant por transacción:
`set_config('app.empresa_actual', <uuid>, true)` + `SET LOCAL ROLE aplicacion`.

Cargar en un PostgreSQL local:
```powershell
cd infraestructura/base-de-datos
.\ejecutar.ps1                 # esquema + semilla + tests
.\ejecutar.ps1 -SoloEsquema    # solo DDL + vistas (producción)
```

## Documentación / Vercel
La documentación visual del modelo vive en [`documentacion/`](documentacion/). Para publicarla
en Vercel, importa el repo y selecciona **Root Directory = `documentacion`** (ver
[documentacion/README.md](documentacion/README.md)).

## Convenciones
UUID como PK · `id_empresa` en toda tabla (multi-tenant) · auditoría
`creado/actualizado/eliminado_en` + `_por_usuario_id` · borrado lógico (`eliminado_en IS NULL`
= activo) · eventos append-only · índices únicos parciales · aislamiento por RLS + FK
compuestas · lenguaje ubicuo del dominio minero.
