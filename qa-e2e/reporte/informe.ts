import fs from "node:fs";
import path from "node:path";
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

const AREAS: Record<string, { titulo: string; explicacion: string }> = {
  seguridad: {
    titulo: "Seguridad",
    explicacion:
      "Comprueba que nadie vea ni toque lo que no es suyo: datos de otra empresa, permisos que no le dieron, o entrar adivinando contrasenas."
  },
  validacion: {
    titulo: "Validacion de datos",
    explicacion:
      "Comprueba que el sistema no acepte datos mal escritos: nombres imposibles, contrasenas debiles, correos falsos."
  },
  contrato: {
    titulo: "Reglas del negocio",
    explicacion:
      "Comprueba que cada accion deje realmente el dato correcto en la base: si dice que guardo, guardo; si dice que rechazo, no guardo nada."
  },
  "casos-uso": {
    titulo: "Uso real en pantalla",
    explicacion: "Recorre el sistema como lo haria una persona, con el navegador, de principio a fin."
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

  async onEnd(resultado: FullResult): Promise<void> {
    const datos = {
      corrida: new Date().toISOString(),
      duracionSegundos: Math.round((Date.now() - this.inicio) / 100) / 10,
      veredicto: resultado.status,
      pruebas: this.pruebas
    };
    fs.mkdirSync(this.carpeta, { recursive: true });
    fs.writeFileSync(
      path.join(this.carpeta, "resultado.json"),
      JSON.stringify(datos, null, 2),
      "utf8"
    );
    fs.writeFileSync(path.join(this.carpeta, "informe.html"), this.html(datos), "utf8");
    console.log(`\nInforme: ${path.join(this.carpeta, "informe.html")}`);
  }

  private motivoDe(resultado: TestResult): string {
    const error = resultado.errors?.[0];
    if (!error) return "";
    const crudo = (error.message ?? "").replace(/\[\d+m/g, "");
    const linea = crudo.split("\n").find((l) => l.trim().length > 0) ?? "";
    return linea.trim().slice(0, 300);
  }

  private html(datos: {
    corrida: string;
    duracionSegundos: number;
    veredicto: string;
    pruebas: Prueba[];
  }): string {
    const porArea = new Map<string, Prueba[]>();
    for (const prueba of datos.pruebas) {
      const lista = porArea.get(prueba.area) ?? [];
      lista.push(prueba);
      porArea.set(prueba.area, lista);
    }
    const total = datos.pruebas.length;
    const pasaron = datos.pruebas.filter((p) => p.estado === "paso").length;
    const fallaron = datos.pruebas.filter((p) => p.estado === "fallo").length;
    const omitidas = datos.pruebas.filter((p) => p.estado === "omitido").length;
    const evaluadas = pasaron + fallaron;
    const porcentaje = evaluadas === 0 ? 0 : Math.round((pasaron / evaluadas) * 100);

    const secciones = [...porArea.entries()]
      .map(([area, lista]) => {
        const meta = AREAS[area] ?? { titulo: area, explicacion: "" };
        const malas = lista.filter((p) => p.estado === "fallo").length;
        const filas = lista
          .map(
            (p) => `<tr class="${p.estado}">
              <td class="marca">${p.estado === "paso" ? "OK" : p.estado === "fallo" ? "FALLA" : "—"}</td>
              <td>${escapar(p.titulo)}${p.motivo ? `<div class="motivo">${escapar(p.motivo)}</div>` : ""}</td>
              <td class="tiempo">${(p.milisegundos / 1000).toFixed(1)}s</td>
            </tr>`
          )
          .join("");
        return `<section>
          <h2>${escapar(meta.titulo)} <span class="cuenta ${malas ? "mal" : "bien"}">${lista.length - malas}/${lista.length}</span></h2>
          <p class="explicacion">${escapar(meta.explicacion)}</p>
          <table>${filas}</table>
        </section>`;
      })
      .join("");

    const resumen =
      fallaron === 0
        ? "Todo lo que se probo funciono como debe."
        : `Hay ${fallaron} ${fallaron === 1 ? "prueba que falla" : "pruebas que fallan"}. Cada una dice abajo que se esperaba.`;

    return `<!doctype html>
<html lang="es"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Calidad · Plataforma Minera</title>
<style>
:root{--fondo:#0e1116;--panel:#171c24;--borde:#252c37;--texto:#e8eaed;--tenue:#9aa4b2;--oro:#d9a440;--bien:#4ec97f;--mal:#ef5f5f}
*{box-sizing:border-box}
body{margin:0;background:var(--fondo);color:var(--texto);font:15px/1.55 "Segoe UI",system-ui,sans-serif;padding:32px 16px}
.hoja{max-width:940px;margin:0 auto}
h1{font-size:26px;margin:0 0 4px}
.fecha{color:var(--tenue);font-size:13px;margin-bottom:24px}
.tarjetas{display:grep;display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:12px;margin-bottom:12px}
.tarjeta{background:var(--panel);border:1px solid var(--borde);border-radius:12px;padding:16px}
.tarjeta .num{font-size:30px;font-weight:600}
.tarjeta .rot{color:var(--tenue);font-size:12px;text-transform:uppercase;letter-spacing:.06em}
.barra{height:10px;border-radius:99px;background:#222835;overflow:hidden;margin:18px 0 6px}
.barra span{display:block;height:100%;background:linear-gradient(90deg,var(--bien),#2f9e5f);width:${porcentaje}%}
.resumen{background:var(--panel);border:1px solid var(--borde);border-left:3px solid var(--oro);border-radius:10px;padding:14px 16px;margin:20px 0 28px}
section{margin-bottom:28px}
h2{font-size:17px;margin:0 0 2px;display:flex;align-items:center;gap:10px}
.cuenta{font-size:12px;padding:2px 8px;border-radius:99px;border:1px solid var(--borde);color:var(--tenue)}
.cuenta.bien{color:var(--bien);border-color:#24503a}
.cuenta.mal{color:var(--mal);border-color:#57282b}
.explicacion{color:var(--tenue);font-size:13px;margin:0 0 10px}
table{width:100%;border-collapse:collapse;background:var(--panel);border:1px solid var(--borde);border-radius:10px;overflow:hidden}
td{padding:9px 12px;border-top:1px solid var(--borde);vertical-align:top}
tr:first-child td{border-top:0}
.marca{width:62px;font-size:11px;font-weight:700;letter-spacing:.04em}
tr.paso .marca{color:var(--bien)}
tr.fallo .marca{color:var(--mal)}
tr.omitido .marca{color:var(--tenue)}
tr.fallo{background:rgba(239,95,95,.06)}
.motivo{color:var(--mal);font-size:12.5px;margin-top:5px;font-family:Consolas,monospace}
.tiempo{width:64px;color:var(--tenue);font-size:12px;text-align:right}
footer{color:var(--tenue);font-size:12px;border-top:1px solid var(--borde);padding-top:14px;margin-top:8px}
</style></head><body><div class="hoja">
<h1>Informe de calidad</h1>
<div class="fecha">Plataforma Minera · ${new Date(datos.corrida).toLocaleString("es-MX")} · ${datos.duracionSegundos}s</div>
<div class="tarjetas">
  <div class="tarjeta"><div class="num">${total}</div><div class="rot">pruebas</div></div>
  <div class="tarjeta"><div class="num" style="color:var(--bien)">${pasaron}</div><div class="rot">pasaron</div></div>
  <div class="tarjeta"><div class="num" style="color:${fallaron ? "var(--mal)" : "var(--tenue)"}">${fallaron}</div><div class="rot">fallaron</div></div>
  <div class="tarjeta"><div class="num" style="color:var(--tenue)">${omitidas}</div><div class="rot">omitidas</div></div>
</div>
<div class="barra"><span></span></div>
<div class="resumen"><strong>${porcentaje}% en verde.</strong> ${escapar(resumen)}</div>
${secciones}
<footer>Cada linea es una promesa que el sistema le hace al cliente. Si dice OK, esa promesa se cumplio contra la base de datos real.</footer>
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
