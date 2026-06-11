# PROYECTO-MINAS-MEX

Sistema de **Planeación y Producción Minera** (mina subterránea polimetálica: Au · Ag · Pb · Zn).
Backend en **Go** sobre **PostgreSQL**, organizado por **capacidades** con DDD ligero y
arquitectura hexagonal, multi-tenant, eventos append-only, auditoría inline y borrado lógico.
Incluye **frontend** servido por el propio backend y **colección Postman** lista para clics.

## Estructura del repositorio

```
PROYECTO-MINAS-MEX/
├── backend/                       Go: DDD hexagonal por capacidades (gobierno, catalogos)
│   ├── pasarela/rutas.go          TODAS las rutas HTTP en un solo lugar (con su permiso)
│   ├── plataforma/                cáscara: web, identidad (JWT), bcrypt, persistencia pgx
│   ├── capacidades/               gobierno (RBAC) y catalogos — soberanía por contratos
│   └── README.md                  endpoints, cómo correr, cómo probar
├── frontend/                      SPA sin build (la sirve el backend en /)
│   ├── index.html · estilos.css · aplicacion.js
│   └── login + panel: usuarios, roles/permisos, minas, empleados, equipos, branding vivo
├── postman/                       Pruebas con uno o varios clics
│   ├── coleccion.json             flujo encadenado completo (login → ids → RBAC en vivo)
│   └── entorno.json               variables (base_url, credenciales, tokens, ids)
├── documentacion/                 Sitio HTML del modelo de datos (se despliega en Vercel)
│   ├── index.html                 portada (10 capacidades)
│   ├── diagrama-general.html      mapa interactivo de las 54 tablas (clic → capacidad)
│   ├── 01..10-*.html              una página por capacidad (incluye 10-seguridad)
│   └── README.md                  cómo desplegar en Vercel
└── infraestructura/
    └── base-de-datos/             Esquema PostgreSQL (fuente de verdad)
        ├── esquema/               00..60 DDL por capacidad + índices + vistas + seguridad RLS
        ├── semilla/20_seed.sql    datos de ejemplo (admin.mina / Mina#2026)
        ├── pruebas/30_tests.sql   42 tests: integración/lógica/gobernanza/aislamiento/RBAC/permisos/seguridad
        ├── pruebas/40_escenario_usuarios.sql   E2E: ciclo de vida real de una empresa y sus usuarios
        ├── pruebas/agentes/       baterías de auditoría senior
        ├── ejecutar.ps1           carga todo en orden
        └── README.md              detalle de la BD
```

## Pruébalo en 2 minutos

```powershell
cd infraestructura/base-de-datos ; ./ejecutar.ps1          # 1) BD local en verde (5433)
cd ../../backend
$env:CADENA_POSTGRES = "postgres://postgres:x@127.0.0.1:5433/mina"
$env:SECRETO_TOKEN   = "secreto-local"
go run ./cmd/servidor                                       # 2) API + frontend en :8080
```

Abre **http://localhost:8080** y entra con `MIN` / `admin.mina` / `Mina#2026` — o importa
la carpeta [`postman/`](postman/) y corre la colección completa con un clic.

> La carpeta `documentacion/` está **separada a propósito** del resto del proyecto para
> poder desplegarla sola en Vercel sin exponer el esquema ni el backend.

## Base de datos
PostgreSQL 17. **54 tablas** en 10 capacidades (schemas: gobierno, catalogos, produccion,
planeacion, reconciliacion, beneficio, estandares, costos, inversiones, seguridad), **19 vistas**
(+1 materializada) y **42 tests + escenario E2E** que pasan en local.

**Seguridad** (capa nueva, datos puros): `tipo_incidente` (catálogo por empresa) e
`incidente` (evento append-only) capturan incidentes y casi-pérdidas en mina —
complementa las demoras de producción sin depender de hardware. **Roadmap documentado,
no implementado:** mantenimiento/órdenes de trabajo y muestreo/grade control. El modelado
geológico 3D queda fuera de alcance a propósito (otra categoría de producto).

**Modelo de administración** (estándar B2B enterprise): la **plataforma** (superadmin)
crea las empresas — con su branding (logo, color, zona horaria, moneda) —; el
**ADMIN_EMPRESA** de cada una da de alta/baja a sus usuarios y les asigna roles
(6 de sistema sembrados + propios), con alcance opcional por mina. Sin auto-registro:
una empresa nunca crea otras empresas. El **catálogo de permisos** es global
(29, convención `recurso.accion`); cada rol lleva su matriz en `rol_permiso` y la vista
`gobierno.v_permisos_usuario` entrega los permisos efectivos para el JWT.

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
