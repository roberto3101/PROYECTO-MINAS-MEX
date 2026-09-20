import { test, expect } from "@playwright/test";
import { PaginaBase } from "../paginas/pagina_base";
import { exigirBaseDesechable, sello, contrasenaValida } from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

test("dar de alta un usuario desde la pantalla lo deja listo para entrar", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  const usuario = sello("qa.uipanel");

  await pantalla.entrarComoEmpresa();
  await pantalla.irA("usuarios");
  await pantalla.abrirFormulario(/Nuevo usuario/);
  await pantalla.llenar("usuario", usuario);
  await pantalla.llenar("nombre", "QA CREADO EN PANTALLA");
  await pantalla.llenar("contrasena", contrasenaValida());
  await pantalla.enviarFormulario();

  await expect(
    pantalla.formularioAbierto,
    "el formulario no se cerro: el alta no se completo"
  ).toBeHidden({ timeout: 15000 });

  const filas = await consultar<{ estado: string; contrasena_hash: string }>(
    "SELECT estado, contrasena_hash FROM gobierno.usuario WHERE usuario = $1",
    [usuario]
  );
  expect(filas, "el usuario creado en pantalla no llego a la base").toHaveLength(1);
  expect(filas[0].estado).toBe("ACTIVO");
  expect(filas[0].contrasena_hash, "la contrasena no quedo cifrada").toMatch(/^\$2[aby]\$/);

  await pantalla.buscar(usuario);
  await expect(pantalla.filaCon(usuario)).toHaveCount(1);
});

test("la pantalla rechaza una contrasena debil y no crea al usuario", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  const usuario = sello("qa.uidebil");

  await pantalla.entrarComoEmpresa();
  await pantalla.irA("usuarios");
  await pantalla.abrirFormulario(/Nuevo usuario/);
  await pantalla.llenar("usuario", usuario);
  await pantalla.llenar("nombre", "QA CLAVE DEBIL");
  await pantalla.llenar("contrasena", "corta1");
  await pantalla.enviarFormulario();

  await pantalla.esperarAviso(/contrasena|contraseña/i);

  const filas = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.usuario WHERE usuario = $1",
    [usuario]
  );
  expect(Number(filas[0].total), "se creo un usuario con contrasena debil").toBe(0);
});

test("el superadmin ve las empresas de la plataforma", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoPlataforma();
  await pantalla.irA("empresas");

  expect(
    await pantalla.filas.count(),
    "el panel de plataforma no muestra ninguna empresa"
  ).toBeGreaterThan(0);
  await expect(
    pantalla.filaCon("MIN").first(),
    "la empresa MIN deberia aparecer en el panel de plataforma"
  ).toBeVisible();
});

test("el superadmin abre una empresa y puede gestionar sus accesos por mina", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoPlataforma();
  await pantalla.irA("empresas");
  await pantalla.abrirDetalleDeFila("MIN");

  await expect(
    pantalla.panelDetalle,
    "el detalle de la empresa no ofrece gestionar los accesos por mina"
  ).toContainText(/Accesos por mina/i);

  await pantalla.panelDetalle.getByRole("button", { name: /Accesos por mina/i }).click();

  await expect(
    pantalla.panelDetalle,
    "al abrir Accesos por mina no se listaron los usuarios de la empresa"
  ).toContainText("admin.mina", { timeout: 15000 });
});
