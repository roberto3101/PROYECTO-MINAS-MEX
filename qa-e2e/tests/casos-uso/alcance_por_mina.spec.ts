import { test, expect } from "@playwright/test";
import { PaginaBase } from "../paginas/pagina_base";
import { comoAdministrador, exigirBaseDesechable, sello, contrasenaValida } from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

async function usuarioConUnaSolaMina(): Promise<{
  usuario: string;
  clave: string;
  nombreMina: string;
}> {
  const administrador = await comoAdministrador();
  const minas = await (await administrador.get("/catalogos/minas?limite=50")).json();
  const elegida = minas.Elementos[0];

  const usuario = sello("qa.pantalla");
  const clave = contrasenaValida();
  const creado = await administrador.post("/gobierno/usuarios", {
    data: { usuario, nombre: "QA UNA MINA", correo: "", contrasena: clave, id_empleado: "" }
  });
  const { id } = await creado.json();

  const rol = await administrador.post("/gobierno/roles", {
    data: { codigo: sello("QA_PANT").toUpperCase(), descripcion: "Rol de pantalla con una mina" }
  });
  const { id: identificadorRol } = await rol.json();
  await administrador.post(`/gobierno/roles/${identificadorRol}/permisos`, {
    data: { permiso: "catalogos.ver" }
  });
  await administrador.post("/gobierno/asignaciones", {
    data: { id_usuario: id, id_rol: identificadorRol, id_mina: elegida.Identificador }
  });
  await administrador.dispose();

  return { usuario, clave, nombreMina: elegida.Nombre };
}

test("el administrador ve el selector con todas las minas de su empresa", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();
  await pantalla.irA("empleados");

  await expect(
    pantalla.selectorDeMina,
    "el selector de mina no aparece en Empleados"
  ).toBeVisible();

  const opciones = await pantalla.selectorDeMina.locator("option").allInnerTexts();
  const minas = await consultar<{ nombre: string }>(
    `SELECT m.nombre FROM catalogos.mina m
     JOIN gobierno.empresa e ON e.id = m.id_empresa
     WHERE e.codigo = 'MIN' AND m.eliminado_en IS NULL AND m.estado = 'ACTIVA'
     ORDER BY m.nombre`
  );

  expect(
    opciones[0],
    "la primera opcion debe permitir ver todas las minas"
  ).toMatch(/Todas/i);
  for (const mina of minas) {
    expect(opciones, `falta la mina ${mina.nombre} en el selector`).toContain(mina.nombre);
  }
});

test("elegir una mina filtra el listado de empleados a esa mina", async ({ page }) => {
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa();
  await pantalla.irA("empleados");

  const conMina = await consultar<{ nombre: string; total: string }>(
    `SELECT m.nombre, count(*)::text AS total
     FROM catalogos.empleado e
     JOIN catalogos.mina m ON m.id = e.id_mina
     JOIN gobierno.empresa em ON em.id = e.id_empresa
     WHERE em.codigo = 'MIN' AND e.eliminado_en IS NULL AND e.estado = 'ACTIVO'
     GROUP BY m.nombre ORDER BY count(*) DESC LIMIT 1`
  );
  test.skip(conMina.length === 0, "la semilla no tiene empleados");

  await pantalla.elegirMina(conMina[0].nombre);

  await expect(
    pantalla.filas,
    `al filtrar por ${conMina[0].nombre} deben quedar solo sus empleados`
  ).toHaveCount(Number(conMina[0].total));

  const minasEnPantalla = await pantalla.filas.evaluateAll((filas) =>
    filas.map((f) => f.textContent ?? "")
  );
  for (const fila of minasEnPantalla) {
    expect(fila, "aparecio un empleado de otra mina tras filtrar").toContain(conMina[0].nombre);
  }
});

test("un usuario con acceso a una sola mina solo ve esa mina, sin poder elegir otra", async ({ page }) => {
  const preparado = await usuarioConUnaSolaMina();
  const pantalla = new PaginaBase(page);
  await pantalla.entrarComoEmpresa(preparado.usuario, preparado.clave);
  await pantalla.irA("minas");

  await expect(
    pantalla.filas,
    "un usuario con una sola mina no debe ver mas minas en el catalogo"
  ).toHaveCount(1);
  await expect(pantalla.filas.first()).toContainText(preparado.nombreMina);

  await pantalla.irA("empleados");
  const opciones = await pantalla.selectorDeMina
    .locator("option")
    .allInnerTexts()
    .catch(() => [] as string[]);

  if (opciones.length > 0) {
    const ajenas = opciones.filter((o) => !/Todas/i.test(o) && o.trim() !== preparado.nombreMina);
    expect(
      ajenas,
      "el selector le ofrece minas a las que no tiene acceso"
    ).toHaveLength(0);
  }
});
