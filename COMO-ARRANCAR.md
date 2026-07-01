# Cómo arrancar el proyecto (un solo clic)

## Requisitos previos (una vez)
- **PostgreSQL 17** instalado (basta la instalación estándar de Windows; el script detecta también 16/15/14).
- **Go** instalado (para compilar el servidor).

No hace falta configurar nada más: el script crea su propio clúster de base de datos local y aislado en la carpeta `.devdb\` del proyecto (no toca ningún PostgreSQL que ya tengas en el puerto 5432).

## Arranque
Haz **doble clic en `ARRANCAR.cmd`**. La primera vez hace todo desde cero:

1. Crea un clúster PostgreSQL local en `.devdb\data` (puerto 5433).
2. Lo levanta.
3. Crea la base de datos `mina`.
4. Carga el esquema, las vistas y los datos de ejemplo leyendo los `.sql` de `infraestructura\base-de-datos`.
5. Compila el servidor Go.
6. Abre el navegador en http://localhost:8080 y deja el servidor corriendo.

Las siguientes veces detecta que ya está instalado y solo levanta la base y el servidor (rápido).

Para **detener** el servidor: `Ctrl+C` en la ventana, o simplemente ciérrala.

## Usuarios de ejemplo
- **Empresa (operación):** empresa `MIN` · usuario `admin.mina` · clave `Mina#2026`
- **Superadmin (plataforma):** usuario `plataforma` · clave `Plataforma#2026`

## Opciones (línea de comandos)
```
ARRANCAR.cmd                 arranca (instala la primera vez)
ARRANCAR.cmd -Reiniciar      borra la base y la recrea desde cero con datos de ejemplo
ARRANCAR.cmd -SoloPreparar   deja todo listo (base + binario) sin abrir el servidor
```
