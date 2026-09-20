import { test, expect } from "@playwright/test";
import { PaginaBase } from "../paginas/pagina_base";
import { exigirBaseDesechable, sello } from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

test("el catalogo de minas muestra las minas de la empresa", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();
  await pantalla.irA("minas");

  const enBase = await consultar<{ total: string }>(
    `SELECT count(*)::text AS total FROM catalogos.mina m
     JOIN gobierno.empresa e ON e.id = m.id_empresa
     WHERE e.codigo = 'MIN' AND m.eliminado_en IS NULL AND m.estado = 'ACTIVA'`
  );

  await expect(
    pantalla.filas,
    "la pantalla de Minas no muestra las minas que hay en la base"
  ).toHaveCount(Number(enBase[0].total));
});

test("dar de alta una mina desde la pantalla la guarda de verdad", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  const nombre = `Mina Pantalla ${sello("")}`;

  await pantalla.entrarComoEmpresa();
  await pantalla.irA("minas");
  await pantalla.abrirFormulario(/Nueva mina/);
  await pantalla.llenar("nombre", nombre);
  await pantalla.llenar("area", "Poniente");
  await pantalla.enviarFormulario();

  await expect(
    pantalla.formularioAbierto,
    "el formulario no se cerro: el alta no se completo"
  ).toBeHidden({ timeout: 15000 });

  const filas = await consultar<{ estado: string }>(
    "SELECT estado FROM catalogos.mina WHERE nombre = $1 AND eliminado_en IS NULL",
    [nombre]
  );
  expect(filas, "la mina creada en pantalla no llego a la base").toHaveLength(1);
  expect(filas[0].estado).toBe("ACTIVA");

  await pantalla.buscar(nombre);
  await expect(
    pantalla.filaCon(nombre),
    "la mina recien creada no aparece en el listado"
  ).toHaveCount(1);
});

test("la pantalla no deja crear una mina sin nombre", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();
  await pantalla.irA("minas");
  await pantalla.abrirFormulario(/Nueva mina/);
  await pantalla.llenar("area", "Sin nombre");
  await pantalla.enviarFormulario();

  await expect(
    pantalla.formularioAbierto,
    "el formulario se cerro pese a faltar el nombre obligatorio"
  ).toBeVisible();
});

test("empleados y equipos muestran datos reales con su mina", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();

  await pantalla.irA("empleados");
  const empleados = await consultar<{ total: string }>(
    `SELECT count(*)::text AS total FROM catalogos.empleado e
     JOIN gobierno.empresa em ON em.id = e.id_empresa
     WHERE em.codigo = 'MIN' AND e.eliminado_en IS NULL AND e.estado = 'ACTIVO'`
  );
  await expect(
    pantalla.filas,
    "la pantalla de Empleados no coincide con lo que hay en la base"
  ).toHaveCount(Number(empleados[0].total));

  await pantalla.irA("equipos");
  await expect(
    pantalla.filas.first(),
    "la pantalla de Equipos quedo vacia"
  ).toBeVisible();
});

test("cada mina se muestra con los datos que el operador necesita", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();
  await pantalla.irA("minas");

  const enBase = await consultar<{ nombre: string; area: string; estado: string }>(
    `SELECT m.nombre, m.area, m.estado FROM catalogos.mina m
     JOIN gobierno.empresa e ON e.id = m.id_empresa
     WHERE e.codigo = 'MIN' AND m.eliminado_en IS NULL AND m.estado = 'ACTIVA'
     ORDER BY m.nombre LIMIT 1`
  );
  test.skip(enBase.length === 0, "la empresa no tiene minas activas");

  const fila = pantalla.filaCon(enBase[0].nombre);
  await expect(fila, "la mina de la base no aparece en pantalla").toHaveCount(1);
  await expect(fila, "la fila no muestra el area de la mina").toContainText(enBase[0].area);
  await expect(fila, "la fila no muestra el estado de la mina").toContainText(enBase[0].estado);
});
