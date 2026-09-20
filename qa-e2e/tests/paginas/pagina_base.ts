import { expect, type Locator, type Page } from "@playwright/test";
import { ADMINISTRADOR, EMPRESA_PRINCIPAL, SUPERADMIN } from "../soporte/base";

export class PaginaBase {
  constructor(protected readonly page: Page) {}

  async abrir(): Promise<void> {
    await this.page.goto("/", { waitUntil: "domcontentloaded" });
    await expect(this.page.locator("#pantalla-login")).toBeVisible();
  }

  async entrarComoEmpresa(
    usuario = ADMINISTRADOR.usuario,
    contrasena = ADMINISTRADOR.contrasena,
    codigoEmpresa = EMPRESA_PRINCIPAL
  ): Promise<void> {
    await this.abrir();
    await this.page.fill("#entrada-codigo-empresa", codigoEmpresa);
    await this.page.fill("#entrada-usuario-empresa", usuario);
    await this.page.fill("#entrada-contrasena-empresa", contrasena);
    await this.page.click("#boton-entrar-empresa");
    await expect(
      this.page.locator("#navegacion .enlace-nav").first(),
      "no se llego al panel despues de entrar"
    ).toBeVisible({ timeout: 15000 });
  }

  async entrarComoPlataforma(): Promise<void> {
    await this.abrir();
    await this.page.click("#pestana-plataforma");
    await this.page.fill("#entrada-usuario-plataforma", SUPERADMIN.usuario);
    await this.page.fill("#entrada-contrasena-plataforma", SUPERADMIN.contrasena);
    await this.page.click("#boton-entrar-plataforma");
    await expect(
      this.page.locator("#navegacion .enlace-nav").first(),
      "no se llego al panel de plataforma"
    ).toBeVisible({ timeout: 15000 });
  }

  async intentarEntrar(usuario: string, contrasena: string, codigoEmpresa = EMPRESA_PRINCIPAL): Promise<void> {
    await this.abrir();
    await this.page.fill("#entrada-codigo-empresa", codigoEmpresa);
    await this.page.fill("#entrada-usuario-empresa", usuario);
    await this.page.fill("#entrada-contrasena-empresa", contrasena);
    await this.page.click("#boton-entrar-empresa");
  }

  get errorDeAcceso(): Locator {
    return this.page.locator("#error-empresa");
  }

  async irA(vista: string): Promise<void> {
    await this.page.click(`.enlace-nav[data-vista="${vista}"]`);
    await expect(this.page.locator("#titulo-vista")).not.toBeEmpty();
    await this.page.waitForLoadState("networkidle");
  }

  get tituloDeVista(): Locator {
    return this.page.locator("#titulo-vista");
  }

  get filas(): Locator {
    return this.page.locator("#vista table.tabla tbody tr");
  }

  filaCon(texto: string): Locator {
    return this.filas.filter({ hasText: texto });
  }

  async buscar(texto: string): Promise<void> {
    const caja = this.page.locator("#vista input").first();
    await caja.fill(texto);
    await this.page.waitForLoadState("networkidle");
  }

  get selectorDeMina(): Locator {
    return this.page.locator("#selector-de-mina select");
  }

  async elegirMina(nombre: string): Promise<void> {
    await this.selectorDeMina.selectOption({ label: nombre });
    await this.page.waitForLoadState("networkidle");
  }

  async abrirFormulario(textoDelBoton: string | RegExp): Promise<void> {
    await this.page.locator("#acciones-vista button", { hasText: textoDelBoton }).first().click();
    await expect(this.page.locator("#capa-modal"), "no se abrio el formulario").toBeVisible();
  }

  async llenar(nombre: string, valor: string): Promise<void> {
    await this.page.locator(`#cuerpo-modal [name="${nombre}"]`).fill(valor);
  }

  async elegir(nombre: string, etiqueta: string): Promise<void> {
    await this.page.locator(`#cuerpo-modal [name="${nombre}"]`).selectOption({ label: etiqueta });
  }

  async elegirPrimeraOpcion(nombre: string): Promise<string> {
    const campo = this.page.locator(`#cuerpo-modal [name="${nombre}"]`);
    const valores = await campo.locator("option").evaluateAll((opciones) =>
      opciones.map((o) => (o as HTMLOptionElement).value).filter((v) => v !== "")
    );
    expect(valores.length, `el selector ${nombre} llego vacio`).toBeGreaterThan(0);
    await campo.selectOption(valores[0]);
    return valores[0];
  }

  async enviarFormulario(): Promise<void> {
    await this.page.locator("#cuerpo-modal button[type=submit]").click();
  }

  get formularioAbierto(): Locator {
    return this.page.locator("#capa-modal");
  }

  async esperarAviso(patron: RegExp): Promise<void> {
    await expect(this.page.locator("#toasts"), "no aparecio el aviso esperado").toContainText(patron, {
      timeout: 15000
    });
  }

  async abrirDetalleDeFila(texto: string): Promise<void> {
    await this.filaCon(texto).first().click();
    await expect(this.page.locator("#panel-detalle"), "no se abrio el panel de detalle").toBeVisible();
  }

  get panelDetalle(): Locator {
    return this.page.locator("#panel-detalle");
  }

  async salir(): Promise<void> {
    await this.page.click("#boton-salir");
    await expect(this.page.locator("#pantalla-login")).toBeVisible({ timeout: 10000 });
  }
}
