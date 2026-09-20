# Suite de calidad — Plataforma Minera

Pruebas automatizadas que comprueban que el sistema **hace de verdad lo que promete**.
No miran si la pantalla "se ve bien": cada prueba ejecuta una accion real contra la API y
despues **va a la base de datos a confirmar el efecto**. Si una prueba dice OK, ese
comportamiento quedo demostrado contra PostgreSQL, no simulado.

## La regla de la casa

> No basta con que algo falle: tiene que fallar **por la razon declarada**.
> Y cuando el sistema rechaza algo, se verifica que **no haya guardado nada**.

## Que se prueba

| Carpeta | Que mide | Ejemplo de pregunta que responde |
|---|---|---|
| `tests/seguridad` | Que nadie vea ni toque lo ajeno | ¿Puede el administrador de una empresa leer usuarios de otra? |
| `tests/validacion` | Que no entren datos mal formados | ¿Acepta una contrasena de 6 letras sin numeros? |
| `tests/contrato` | Que cada accion deje el dato correcto | ¿Al crear un usuario queda con su contrasena cifrada y en su empresa? |
| `tests/casos-uso` | El sistema usado por una persona | Recorrido por navegador, de principio a fin |

## Requisitos

- El proyecto preparado (usa el `ARRANCAR.cmd` de la raiz).
- Node 20+.

## Preparar y correr

1. Crear la base **desechable** y compilar (desde la raiz del proyecto):

```bash
powershell -ExecutionPolicy Bypass -File arrancar-todo.ps1 -Base mina_qa -SoloPreparar
```

2. Levantar el servidor apuntado a esa base, en el puerto 8099:

```bash
CADENA_POSTGRES="postgres://postgres:x@127.0.0.1:5433/mina_qa" SECRETO_TOKEN="secreto-qa" DIRECCION=":8099" DIRECTORIO_FRONTEND="./frontend" DIRECTORIO_ARCHIVOS="./datos" .devdb/servidor.exe
```

3. Correr las pruebas:

```bash
cd qa-e2e && npm install && npm run probar
```

El informe queda en `qa-e2e/reporte/informe.html`.

## Candado de seguridad

Las pruebas **escriben**. Por eso exigen que la base se llame `mina_qa` (o contenga `qa`):
si apuntas a la base real, la suite se niega a correr. Se controla con `BASE_QA` y
`CADENA_QA`.

## Atajos

```bash
npm run probar:seguridad     # solo seguridad
npm run probar:validacion    # solo validacion de datos
npm run probar:contrato      # solo reglas de negocio
VER_NAVEGADOR=1 npm run probar:casos-uso   # con el navegador visible
```

## Nota sobre esta plataforma

En otros sistemas la logica vive en *stored procedures* y se valida contra ellos. Aqui no
hay SPs: la logica vive en Go y la base aporta el blindaje (RLS por empresa, llaves
foraneas compuestas y vistas). Por eso estas pruebas validan la costura completa
**API -> Go -> PostgreSQL**, que es el equivalente exacto.

## Que produce cada corrida

En `reporte/` quedan tres versiones del mismo informe:

| Archivo | Para que sirve |
|---|---|
| `informe.html` | Leerlo en pantalla, con colores y explicaciones en lenguaje llano |
| `informe.pdf` | Enviarlo al cliente tal cual |
| `informe.xlsx` | Filtrarlo en Excel: una hoja por area mas el resumen |

## Hallazgos que ya encontro

1. **El alcance por mina no viajaba en el token de sesion.** El login calculaba a que minas
   podia entrar cada usuario, pero al emitir el token no se copiaban esos datos. Resultado:
   las pantallas de Minas, Empleados y Equipos salian **vacias para todo el mundo**.
   Corregido en `manejador_gobierno.go`.
2. **El superadmin no podia abrir ninguna empresa.** El detalle contaba las minas desde el
   esquema `catalogos` usando el rol `plataforma`, que por diseño no tiene acceso ahi, y la
   respuesta era siempre un error 500. Como el boton "Accesos por mina" vive dentro de ese
   panel, **toda la gestion de accesos era inalcanzable**. Corregido contando las minas
   mediante impersonacion del tenant, sin ampliar privilegios.
3. **Los errores internos de la base se enviaban al cliente** (incluido el texto y el
   SQLSTATE de PostgreSQL). Ahora se registran en el servidor y al cliente le llega un
   mensaje neutro.
4. **Una empresa inexistente respondia 500 en vez de 404.**

## Carencias detectadas (no son fallos, es funcionalidad que no existe)

- El catalogo de **Minas no abre ficha de detalle** al pulsar una fila, aunque el backend
  ya expone `GET /catalogos/minas/{id}`. Usuarios y Empresas si lo hacen.
- Un usuario con alcance limitado **puede dar de alta** un empleado o equipo en una mina
  fuera de su alcance: el filtro de lectura existe, el de escritura no.
