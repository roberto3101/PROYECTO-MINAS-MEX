# Backend — Sistema de Planeación y Producción Minera

Go + PostgreSQL. **DDD modular por capacidades** con arquitectura hexagonal y
**soberanía por contratos**. Capacidades construidas: **gobierno** (acceso, RBAC,
multi-tenant) y **catalogos** (minas, empleados, equipos). Lenguaje ubicuo en español;
el propio código explica lo que hace. El servidor además **sirve el frontend** (`/frontend`)
en `/` — un solo proceso para probar todo.

## Reglas de arquitectura

1. **Soberanía de capacidades.** Ninguna capacidad consulta las tablas de otra. Si
   `produccion` necesita saber los permisos de un usuario, llama al **contrato público**
   de gobierno (`capacidades/gobierno/contrato`), nunca a `gobierno.usuario` directamente.
   La implementación interna de cada capacidad queda oculta tras su contrato.

2. **Multi-tenant en la base, no en el código.** El backend nunca escribe `WHERE
   id_empresa`. La unidad de trabajo abre la transacción, fija el tenant y baja de rol:

   ```
   BEGIN;
   SELECT set_config('app.empresa_actual', $tenant, true);
   SET LOCAL ROLE aplicacion;
   ... consultas normales ...
   COMMIT;
   ```

   El aislamiento lo garantizan RLS + FKs compuestas de la base (ver `infraestructura/`).

3. **Cáscara vs inteligencia.** `plataforma/` es infraestructura transversal (entrada HTTP,
   pasarela, identidad, contexto de tenant, persistencia). `capacidades/` son los dominios
   de negocio. Las capacidades consumen la cáscara; la cáscara no conoce las capacidades.

4. **Sin comentarios.** Los nombres (tipos, funciones, variables) cargan el significado.

5. **Rutas centralizadas.** TODAS las rutas HTTP viven en un único archivo:
   [`pasarela/rutas.go`](pasarela/rutas.go). Las capacidades exponen manejadores; la
   pasarela decide rutas y permisos. Cero rutas sueltas por el código.

## Estructura

```
backend/
├── cmd/servidor/main.go              cableado e inicio del servidor
├── pasarela/rutas.go                 TODAS las rutas + permiso exigido (único lugar)
├── compartido/                       núcleo compartido (identificador UUID, reloj)
├── plataforma/                       CÁSCARA (infraestructura transversal)
│   ├── contexto/                     tenant en el context.Context
│   ├── identidad/                    emisor/verificador de token JWT (HS256), sesion
│   ├── seguridad/                    cifrador de contraseña (bcrypt)
│   ├── persistencia/                 pool pgx, unidad de trabajo (tenant) y de plataforma
│   └── entrada/web/                  servidor HTTP, autenticador, respuestas
└── capacidades/{gobierno, catalogos} CAPACIDADES de negocio, cada una con:
    ├── dominio/                      entidades + invariantes — Go puro, sin dependencias
    ├── puertos/                      interfaces que la capacidad necesita (repos, lector)
    ├── aplicacion/                   casos de uso (una operación = un archivo)
    ├── infraestructura/              adaptadores PostgreSQL + implementación del contrato
    ├── contrato/                     cara pública de la capacidad (lo que otras consumen)
    └── entrada/                      manejadores HTTP (sin rutas: las pone la pasarela)
```

## Casos de uso

**gobierno:** `iniciar_sesion` · `registrar_usuario` (con contraseña) · `desactivar_usuario` ·
`listar_usuarios` · `crear_rol` · `listar_roles` · `listar_permisos` · `conceder_permiso_a_rol` ·
`asignar_rol` · `revocar_rol` · `listar_asignaciones_de_usuario` · `configurar_empresa`.

**catalogos:** `crear_mina` · `listar_minas` · `contratar_empleado` · `listar_empleados` ·
`dar_de_alta_equipo` · `listar_equipos` · `listar_tipos_de_equipo` · `listar_modulos_de_trabajo`.

## Endpoints (declarados en `pasarela/rutas.go`)

| Método y ruta | Permiso exigido | Operación |
|---|---|---|
| `GET /salud` | — | Sonda de vida |
| `POST /sesiones` | — | Inicia sesión y emite el JWT con los permisos |
| `GET /gobierno/empresa` | (autenticado) | Empresa de la sesión con su branding |
| `PUT /gobierno/empresa` | `empresa.configurar` | Configura logo, color, zona horaria, moneda |
| `GET /gobierno/sesion/permisos` | (autenticado) | Permisos efectivos del usuario de la sesión |
| `GET /gobierno/usuarios` | `usuarios.ver` | Lista los usuarios de la empresa |
| `POST /gobierno/usuarios` | `usuarios.crear` | Alta de usuario (con contraseña bcrypt) |
| `DELETE /gobierno/usuarios/{id}` | `usuarios.desactivar` | Baja lógica del usuario |
| `GET /gobierno/usuarios/{id}/asignaciones` | `roles.ver` | Asignaciones (vigentes y revocadas) |
| `GET /gobierno/roles` | `roles.ver` | Roles con su matriz de permisos |
| `POST /gobierno/roles` | `roles.crear` | Crea un rol propio |
| `POST /gobierno/roles/{id}/permisos` | `roles.editar` | Concede un permiso a un rol propio |
| `GET /gobierno/permisos` | `roles.ver` | Catálogo global de permisos (29) |
| `POST /gobierno/asignaciones` | `roles.asignar` | Asigna rol (alcance opcional por mina) |
| `DELETE /gobierno/asignaciones/{id}` | `roles.asignar` | Revoca una asignación (trazable) |
| `GET /catalogos/minas` | `catalogos.ver` | Lista minas |
| `POST /catalogos/minas` | `catalogos.editar` | Crea una mina |
| `GET /catalogos/empleados` | `catalogos.ver` | Lista empleados (con su mina) |
| `POST /catalogos/empleados` | `catalogos.editar` | Contrata un empleado |
| `GET /catalogos/equipos` | `catalogos.ver` | Lista equipos (tipo, módulo, mina) |
| `POST /catalogos/equipos` | `catalogos.editar` | Da de alta un equipo (incl. TIRO/BANDA) |
| `GET /catalogos/tipos-de-equipo` | `catalogos.ver` | Lookup para selects |
| `GET /catalogos/modulos-de-trabajo` | `catalogos.ver` | Lookup para selects |
| `GET /` | — | **Frontend** (SPA servida desde `/frontend`) |

## Cómo correr (y probar tú mismo)

```powershell
$env:CADENA_POSTGRES = "postgres://postgres:x@127.0.0.1:5433/mina"
$env:SECRETO_TOKEN   = "cambiar-en-produccion"
go run ./cmd/servidor
```

Abre **http://localhost:8080** → frontend con login.
**Credenciales demo (del seed):** empresa `MIN` · usuario `admin.mina` · contraseña `Mina#2026`.

**Postman:** importa `../postman/coleccion.json` + `../postman/entorno.json` y dale
**Run** a la colección completa: el login guarda el token solo, cada creación captura su id,
y el flujo demuestra el RBAC en vivo (el operador nuevo ve catálogos pero recibe 403 en
gobierno; al revocarle el rol entra con 0 permisos; al desactivarlo, 401).

## Pruebas

```powershell
go test ./...                       # unitarias (dominio, casos de uso, token) — sin BD
$env:CADENA_POSTGRES = "postgres://postgres:x@127.0.0.1:5433/mina"
go test -tags integracion ./pruebas/integracion/   # contra la BD viva (tenant + RLS + login)
```

La prueba de integración verifica el flujo completo: inicio de sesión bajo el rol
`plataforma`, alta de usuario bajo el rol `aplicacion` con el tenant fijado, y que el
**aislamiento RLS** impide que otra empresa vea ese usuario.
