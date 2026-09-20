import fs from "node:fs";
import path from "node:path";
import { chromium } from "@playwright/test";
import ExcelJS from "exceljs";
import type {
  FullConfig,
  FullResult,
  Reporter,
  Suite,
  TestCase,
  TestResult
} from "@playwright/test/reporter";

type Prueba = {
  area: string;
  archivo: string;
  titulo: string;
  estado: "paso" | "fallo" | "omitido";
  milisegundos: number;
  motivo: string;
};

type Datos = {
  corrida: string;
  duracionSegundos: number;
  pruebas: Prueba[];
};

const AREAS: Record<string, { titulo: string; pregunta: string; explicacion: string }> = {
  seguridad: {
    titulo: "Candados",
    pregunta: "¿Puede alguien ver o tocar lo que no es suyo?",
    explicacion:
      "Como las llaves de una casa: cada empresa tiene la suya y no abre la puerta del vecino. Aqui probamos a forzar esas puertas."
  },
  validacion: {
    titulo: "Datos bien escritos",
    pregunta: "¿Deja el sistema guardar datos mal hechos?",
    explicacion:
      "Si alguien escribe un correo que no existe o una contrasena facil de adivinar, el sistema tiene que decir que no."
  },
  contrato: {
    titulo: "Promesas cumplidas",
    pregunta: "Cuando el sistema dice que guardo algo, ¿de verdad lo guardo?",
    explicacion:
      "Despues de cada accion vamos a la base de datos a mirar con nuestros propios ojos que el dato quedo bien."
  },
  "casos-uso": {
    titulo: "Uso de verdad",
    pregunta: "¿Funciona cuando una persona lo usa con el raton?",
    explicacion:
      "Abrimos el programa en un navegador de verdad, entramos, damos clics y comprobamos lo que aparece en pantalla."
  }
};

export default class InformeDeCalidad implements Reporter {
  private readonly pruebas: Prueba[] = [];
  private inicio = Date.now();
  private carpeta = "";

  onBegin(config: FullConfig, _suite: Suite): void {
    this.inicio = Date.now();
    this.carpeta = path.join(path.dirname(config.configFile ?? process.cwd()), "reporte");
  }

  onTestEnd(test: TestCase, resultado: TestResult): void {
    const relativo = path.relative(path.join(this.carpeta, "..", "tests"), test.location.file);
    const area = relativo.split(path.sep)[0] ?? "otros";
    const estado =
      resultado.status === "passed" ? "paso" : resultado.status === "skipped" ? "omitido" : "fallo";
    this.pruebas.push({
      area,
      archivo: path.basename(test.location.file),
      titulo: test.title,
      estado,
      milisegundos: resultado.duration,
      motivo: this.motivoDe(resultado)
    });
  }

  async onEnd(_resultado: FullResult): Promise<void> {
    const datos: Datos = {
      corrida: new Date().toISOString(),
      duracionSegundos: Math.round((Date.now() - this.inicio) / 100) / 10,
      pruebas: this.pruebas
    };
    fs.mkdirSync(this.carpeta, { recursive: true });

    const html = this.html(datos);
    const rutaHtml = path.join(this.carpeta, "informe.html");
    fs.writeFileSync(rutaHtml, html, "utf8");
    fs.writeFileSync(path.join(this.carpeta, "resultado.json"), JSON.stringify(datos, null, 2), "utf8");

    await this.excel(datos);
    await this.pdf(html);

    console.log(`\nInforme para leer:   ${rutaHtml}`);
    console.log(`Informe para enviar: ${path.join(this.carpeta, "informe.pdf")}`);
    console.log(`Informe para Excel:  ${path.join(this.carpeta, "informe.xlsx")}`);
  }

  private motivoDe(resultado: TestResult): string {
    const error = resultado.errors?.[0];
    if (!error) return "";
    const crudo = (error.message ?? "").replace(/\[\d+m/g, "");
    const linea = crudo.split("\n").find((l) => l.trim().length > 0) ?? "";
    return linea.trim().slice(0, 300);
  }

  private agrupar(datos: Datos): Map<string, Prueba[]> {
    const porArea = new Map<string, Prueba[]>();
    for (const prueba of datos.pruebas) {
      const lista = porArea.get(prueba.area) ?? [];
      lista.push(prueba);
      porArea.set(prueba.area, lista);
    }
    return porArea;
  }

  private async excel(datos: Datos): Promise<void> {
    const libro = new ExcelJS.Workbook();
    libro.creator = "Plataforma Minera";
    libro.created = new Date();

    const pasaron = datos.pruebas.filter((p) => p.estado === "paso").length;
    const fallaron = datos.pruebas.filter((p) => p.estado === "fallo").length;

    const resumen = libro.addWorksheet("Resumen");
    resumen.columns = [
      { header: "Concepto", key: "concepto", width: 46 },
      { header: "Valor", key: "valor", width: 22 }
    ];
    resumen.addRows([
      { concepto: "Fecha de la revision", valor: new Date(datos.corrida).toLocaleString("es-MX") },
      { concepto: "Pruebas realizadas", valor: datos.pruebas.length },
      { concepto: "Pruebas que pasaron", valor: pasaron },
      { concepto: "Pruebas que fallaron", valor: fallaron },
      { concepto: "Duracion (segundos)", valor: datos.duracionSegundos },
      { concepto: "Veredicto", valor: fallaron === 0 ? "TODO EN ORDEN" : "REQUIERE ATENCION" }
    ]);
    resumen.getRow(1).font = { bold: true };

    for (const [area, lista] of this.agrupar(datos)) {
      const meta = AREAS[area] ?? { titulo: area, pregunta: "", explicacion: "" };
      const hoja = libro.addWorksheet(meta.titulo.slice(0, 28));
      hoja.columns = [
        { header: "Resultado", key: "estado", width: 12 },
        { header: "Que se comprobo", key: "titulo", width: 82 },
        { header: "Si fallo, por que", key: "motivo", width: 60 },
        { header: "Segundos", key: "segundos", width: 10 }
      ];
      hoja.getRow(1).font = { bold: true };
      for (const prueba of lista) {
        const fila = hoja.addRow({
          estado: prueba.estado === "paso" ? "OK" : prueba.estado === "fallo" ? "FALLA" : "omitida",
          titulo: prueba.titulo,
          motivo: prueba.motivo,
          segundos: Math.round(prueba.milisegundos / 100) / 10
        });
        fila.getCell("estado").font = {
          bold: true,
          color: { argb: prueba.estado === "paso" ? "FF1E7F44" : "FFC62828" }
        };
      }
    }

    await libro.xlsx.writeFile(path.join(this.carpeta, "informe.xlsx"));
  }

  private async pdf(html: string): Promise<void> {
    try {
      const navegador = await chromium.launch();
      const pagina = await navegador.newPage();
      await pagina.setContent(html, { waitUntil: "load" });
      await pagina.pdf({
        path: path.join(this.carpeta, "informe.pdf"),
        format: "A4",
        printBackground: true,
        margin: { top: "14mm", bottom: "14mm", left: "12mm", right: "12mm" }
      });
      await navegador.close();
    } catch (error) {
      console.log(`No se pudo generar el PDF: ${(error as Error).message}`);
    }
  }

  private html(datos: Datos): string {
    const porArea = this.agrupar(datos);
    const total = datos.pruebas.length;
    const pasaron = datos.pruebas.filter((p) => p.estado === "paso").length;
    const fallaron = datos.pruebas.filter((p) => p.estado === "fallo").length;
    const omitidas = datos.pruebas.filter((p) => p.estado === "omitido").length;
    const evaluadas = pasaron + fallaron;
    const porcentaje = evaluadas === 0 ? 0 : Math.round((pasaron / evaluadas) * 100);
    const todoBien = fallaron === 0;

    const secciones = [...porArea.entries()]
      .map(([area, lista]) => {
        const meta = AREAS[area] ?? { titulo: area, pregunta: "", explicacion: "" };
        const malas = lista.filter((p) => p.estado === "fallo").length;
        const filas = lista
          .map(
            (p) => `<tr class="${p.estado}">
              <td class="marca">${p.estado === "paso" ? "✓" : p.estado === "fallo" ? "✕" : "–"}</td>
              <td>${escapar(p.titulo)}${p.motivo ? `<div class="motivo">${escapar(p.motivo)}</div>` : ""}</td>
            </tr>`
          )
          .join("");
        return `<section>
          <div class="cabecera-area">
            <h2>${escapar(meta.titulo)}</h2>
            <span class="cuenta ${malas ? "mal" : "bien"}">${lista.length - malas} de ${lista.length}</span>
          </div>
          <p class="pregunta">${escapar(meta.pregunta)}</p>
          <p class="explicacion">${escapar(meta.explicacion)}</p>
          <table>${filas}</table>
        </section>`;
      })
      .join("");

    return `<!doctype html>
<html lang="es"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Revision de calidad · Plataforma Minera</title>
<style>
:root{--fondo:#0d1017;--panel:#161b23;--borde:#242b36;--texto:#e9ecf1;--tenue:#98a2b3;--oro:#d9a440;--bien:#4ec97f;--mal:#ef5f5f}
*{box-sizing:border-box}
body{margin:0;background:var(--fondo);color:var(--texto);font:15px/1.6 "Segoe UI",system-ui,sans-serif;padding:34px 16px}
.hoja{max-width:900px;margin:0 auto}
.marca-agua{display:flex;align-items:center;gap:9px;color:var(--oro);font-weight:600;letter-spacing:.14em;font-size:11px;text-transform:uppercase;margin-bottom:22px}
.marca-agua i{width:22px;height:2px;background:var(--oro);display:block}
h1{font-size:30px;margin:0 0 6px;letter-spacing:-.4px}
.fecha{color:var(--tenue);font-size:13px;margin-bottom:26px}
.veredicto{display:flex;align-items:center;gap:18px;background:var(--panel);border:1px solid var(--borde);border-left:4px solid ${todoBien ? "var(--bien)" : "var(--mal)"};border-radius:14px;padding:22px 24px;margin-bottom:26px}
.veredicto .icono{font-size:38px;line-height:1}
.veredicto h2{margin:0 0 3px;font-size:21px}
.veredicto p{margin:0;color:var(--tenue);font-size:14px}
.tarjetas{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:12px;margin-bottom:14px}
.tarjeta{background:var(--panel);border:1px solid var(--borde);border-radius:12px;padding:15px 16px}
.tarjeta .num{font-size:28px;font-weight:650;line-height:1.15}
.tarjeta .rot{color:var(--tenue);font-size:11px;text-transform:uppercase;letter-spacing:.07em;margin-top:3px}
.barra{height:9px;border-radius:99px;background:#1e242e;overflow:hidden;margin:16px 0 30px}
.barra span{display:block;height:100%;background:linear-gradient(90deg,var(--oro),var(--bien));width:${porcentaje}%}
section{margin-bottom:30px;break-inside:avoid}
.cabecera-area{display:flex;align-items:center;gap:12px;margin-bottom:2px}
h2{font-size:18px;margin:0}
.cuenta{font-size:11.5px;padding:3px 10px;border-radius:99px;border:1px solid var(--borde);color:var(--tenue);white-space:nowrap}
.cuenta.bien{color:var(--bien);border-color:#22503a}
.cuenta.mal{color:var(--mal);border-color:#57282b}
.pregunta{margin:0 0 4px;font-size:15px;color:var(--oro)}
.explicacion{color:var(--tenue);font-size:13.5px;margin:0 0 12px;max-width:70ch}
table{width:100%;border-collapse:collapse;background:var(--panel);border:1px solid var(--borde);border-radius:12px;overflow:hidden}
td{padding:10px 13px;border-top:1px solid var(--borde);vertical-align:top}
tr:first-child td{border-top:0}
.marca{width:34px;font-size:15px;font-weight:700;text-align:center}
tr.paso .marca{color:var(--bien)}
tr.fallo .marca{color:var(--mal)}
tr.omitido .marca{color:var(--tenue)}
tr.fallo{background:rgba(239,95,95,.07)}
.motivo{color:var(--mal);font-size:12.5px;margin-top:6px;font-family:Consolas,monospace}
footer{color:var(--tenue);font-size:12.5px;border-top:1px solid var(--borde);padding-top:16px;margin-top:10px;max-width:70ch}
</style></head><body><div class="hoja">
<div class="marca-agua"><i></i> Plataforma Minera · Revision de calidad</div>
<h1>${todoBien ? "Todo funciona como debe" : `Hay ${fallaron} ${fallaron === 1 ? "cosa" : "cosas"} por arreglar`}</h1>
<div class="fecha">${new Date(datos.corrida).toLocaleString("es-MX")} · la revision tardo ${datos.duracionSegundos} segundos</div>

<div class="veredicto">
  <div class="icono">${todoBien ? "✓" : "!"}</div>
  <div>
    <h2>${pasaron} de ${evaluadas} comprobaciones salieron bien</h2>
    <p>${
      todoBien
        ? "Probamos a romper el sistema de todas las formas que se nos ocurrieron y aguanto."
        : "Abajo se explica, en la seccion correspondiente, que fallo y por que."
    }</p>
  </div>
</div>

<div class="tarjetas">
  <div class="tarjeta"><div class="num">${total}</div><div class="rot">comprobaciones</div></div>
  <div class="tarjeta"><div class="num" style="color:var(--bien)">${pasaron}</div><div class="rot">bien</div></div>
  <div class="tarjeta"><div class="num" style="color:${fallaron ? "var(--mal)" : "var(--tenue)"}">${fallaron}</div><div class="rot">mal</div></div>
  <div class="tarjeta"><div class="num" style="color:var(--tenue)">${omitidas}</div><div class="rot">sin datos</div></div>
</div>
<div class="barra"><span></span></div>

${secciones}

<footer>
Como leer esto: cada linea con <strong>✓</strong> es una promesa que el sistema le hace a quien lo compra, y que
comprobamos de verdad — no preguntandole al programa si hizo su trabajo, sino yendo a la base de datos a mirarlo.
Cuando el sistema rechaza algo, ademas verificamos que no haya guardado nada a escondidas.
</footer>
</div></body></html>`;
  }
}

function escapar(texto: string): string {
  return texto
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
