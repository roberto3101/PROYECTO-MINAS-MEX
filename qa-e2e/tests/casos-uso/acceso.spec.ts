import { test, expect } from "@playwright/test";
import { PaginaBase } from "../paginas/pagina_base";
import { exigirBaseDesechable, sello, contrasenaValida, comoAdministrador } from "../soporte/base";

test.beforeAll(() => exigirBaseDesechable());

test("una persona entra con sus credenciales y ve su empresa y sus permisos", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();

  await expect(
    page.locator("#usuario-sesion"),
    "el panel debe decir con que usuario se entro"
  ).toContainText("admin.mina");
  await expect(
    page.locator("#lateral"),
    "el panel debe mostrar la empresa a la que se entro"
  ).toContainText("Minera Ejemplo");
  await expect(pantalla.tituloDeVista).toContainText("Panel");
});

test("con la contrasena equivocada no entra y la pantalla lo explica", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.intentarEntrar("admin.mina", "ClaveQueNoEs123");

  await expect(
    pantalla.errorDeAcceso,
    "la pantalla debe decir que las credenciales no son correctas"
  ).toContainText(/incorrect/i, { timeout: 15000 });
  await expect(
    page.locator("#pantalla-login"),
    "con credenciales malas no puede pasar al panel"
  ).toBeVisible();
});

test("un usuario sin permisos entra pero no ve ninguna seccion de gestion", async ({ page }) => {
  const administrador = await comoAdministrador();
  const usuario = sello("qa.pelado");
  const clave = contrasenaValida();
  const creado = await administrador.post("/gobierno/usuarios", {
    data: { usuario, nombre: "QA SIN ACCESO", correo: "", contrasena: clave, id_empleado: "" }
  });
  expect(creado.status()).toBe(201);
  await administrador.dispose();

  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa(usuario, clave);

  const secciones = await page.locator("#navegacion .enlace-nav").evaluateAll((enlaces) =>
    enlaces.map((e) => (e as HTMLElement).dataset.vista)
  );
  const GESTION = ["usuarios", "roles", "minas", "empleados", "equipos", "empresas"];
  expect(
    secciones.filter((vista) => GESTION.includes(vista ?? "")),
    "un usuario sin permisos no debe ver secciones de gestion en el menu"
  ).toHaveLength(0);
  expect(secciones, "todo usuario autenticado conserva su panel").toContain("panel");
});

test("al salir, la sesion se cierra y vuelve a pedir credenciales", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();
  await pantalla.salir();

  await page.reload();
  await expect(
    page.locator("#pantalla-login"),
    "tras salir y recargar no debe quedar sesion abierta"
  ).toBeVisible();
});
