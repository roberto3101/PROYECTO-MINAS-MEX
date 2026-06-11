# Backend — Sistema de Planeación y Producción Minera

Go + PostgreSQL. **DDD modular por capacidades** con arquitectura hexagonal y
**soberanía por contratos**. Primera capacidad construida: **gobierno** (acceso, RBAC,
multi-tenant). Lenguaje ubicuo en español; el propio código explica lo que hace.

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

## Estructura

```
backend/
├── cmd/servidor/main.go              cableado e inicio del servidor
├── compartido/                       núcleo compartido (identificador UUID, reloj)
├── plataforma/                       CÁSCARA (infraestructura transversal)
│   ├── contexto/                     tenant en el context.Context
│   ├── identidad/                    emisor/verificador de token JWT (HS256), sesion
│   ├── seguridad/                    cifrador de contraseña (bcrypt)
│   ├── persistencia/                 pool pgx, unidad de trabajo (tenant) y de plataforma
│   └── entrada/web/                  servidor HTTP, autenticador, respuestas
└── capacidades/gobierno/             CAPACIDAD de gobierno y acceso
    ├── dominio/                      entidades + invariantes (empresa, usuario, rol,
    │                                 permiso, asignacion_rol) — Go puro, sin dependencias
    ├── puertos/                      interfaces que el dominio necesita (repositorios,
    │                                 unidad de trabajo, cifrador, lector de acceso)
    ├── aplicacion/                   casos de uso (una operación = un archivo)
    ├── infraestructura/              repositorios PostgreSQL + implementación del contrato
    ├── contrato/                     cara pública de la capacidad (lo que otras consumen)
    └── entrada/                      manejador HTTP de la capacidad
```

## Casos de uso (capacidad gobierno)

`iniciar_sesion` · `registrar_usuario` · `desactivar_usuario` · `crear_rol` ·
`conceder_permiso_a_rol` · `asignar_rol` · `revocar_rol` · `configurar_empresa`.

## Endpoints

| Método y ruta | Permiso exigido | Operación |
|---|---|---|
| `POST /sesiones` | — | Inicia sesión (empresa+usuario+contraseña) y emite el JWT |
| `POST /gobierno/usuarios` | `usuarios.crear` | Da de alta un usuario en la empresa de la sesión |
| `DELETE /gobierno/usuarios/{id}` | `usuarios.desactivar` | Baja lógica del usuario |
| `POST /gobierno/roles` | `roles.crear` | Crea un rol propio de la empresa |
| `POST /gobierno/roles/{id}/permisos` | `roles.editar` | Concede un permiso a un rol propio |
| `POST /gobierno/asignaciones` | `roles.asignar` | Asigna un rol a un usuario (alcance opcional por mina) |
| `DELETE /gobierno/asignaciones/{id}` | `roles.asignar` | Revoca una asignación (baja lógica trazable) |
| `PUT /gobierno/empresa` | `empresa.configurar` | Configura el branding de la empresa |
| `GET /gobierno/permisos-vigentes` | (autenticado) | Permisos efectivos del usuario de la sesión |
| `GET /salud` | — | Sonda de vida |

## Cómo correr

```powershell
$env:CADENA_POSTGRES = "postgres://postgres:x@127.0.0.1:5433/mina"
$env:SECRETO_TOKEN   = "cambiar-en-produccion"
$env:DIRECCION       = ":8080"
go run ./cmd/servidor
```

## Pruebas

```powershell
go test ./...                       # unitarias (dominio, casos de uso, token) — sin BD
$env:CADENA_POSTGRES = "postgres://postgres:x@127.0.0.1:5433/mina"
go test -tags integracion ./pruebas/integracion/   # contra la BD viva (tenant + RLS + login)
```

La prueba de integración verifica el flujo completo: inicio de sesión bajo el rol
`plataforma`, alta de usuario bajo el rol `aplicacion` con el tenant fijado, y que el
**aislamiento RLS** impide que otra empresa vea ese usuario.
