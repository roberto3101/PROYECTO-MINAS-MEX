# Documentación — Sistema de Planeación y Producción Minera

Sitio estático (HTML/CSS/JS, sin build) que documenta el modelo de datos por capacidades.
Punto de entrada: **`index.html`**.

## Páginas
- `index.html` — portada con las 7 capacidades
- `01-catalogos-produccion.html` … `07-inversiones.html` — una por capacidad
- `estilos.css`, `scripts.js` — compartidos (Mermaid + tooltips SQL)

## Desplegar en Vercel
Esta carpeta es autónoma; despliégala sin tocar el resto del repo:

1. En Vercel → **Add New → Project** → importa el repo `PROYECTO-MINAS-MEX`.
2. En **Root Directory**, selecciona **`documentacion`** (¡importante! así solo se publica
   la documentación y no el resto del proyecto).
3. Framework Preset: **Other** (es estático, no necesita build).
4. Deploy. La home servirá `index.html`.

El `vercel.json` de esta carpeta solo añade cabeceras de caché; no requiere build.
