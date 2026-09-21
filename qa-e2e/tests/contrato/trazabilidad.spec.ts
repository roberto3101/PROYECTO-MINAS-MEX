import { test, expect } from "@playwright/test";
import { comoAdministrador, exigirBaseDesechable, sello, contrasenaValida } from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

async function identificadorDeAdministrador(): Promise<string> {
  const filas = await consultar<{ id: string }>(
    `SELECT u.id FROM gobierno.usuario u
     JOIN gobierno.empresa e ON e.id = u.id_empresa
     WHERE e.codigo = 'MIN' AND u.usuario = 'admin.mina'`
  );
  expect(filas, "no se encontro al administrador de la semilla").toHaveLength(1);
  return filas[0].id;
}

test("lo que se da de alta queda con el nombre de quien lo capturo", async () => {
  const administrador = await comoAdministrador();
  const nombre = `Mina Autor ${sello("")}`;

  const creada = await administrador.post("/catalogos/minas", {
    data: { nombre, area: "Norte", niveles: "" }
  });
  expect(creada.status()).toBe(201);
  const { id } = await creada.json();

  const filas = await consultar<{ creado_por: string | null; creado_en: string }>(
    "SELECT creado_por_usuario_id::text AS creado_por, creado_en::text FROM catalogos.mina WHERE id = $1",
    [id]
  );
  expect(
    filas[0].creado_por,
    "la mina quedo sin autor: el historico no sirve para auditar"
  ).toBe(await identificadorDeAdministrador());
  expect(filas[0].creado_en, "la mina quedo sin fecha de alta").toBeTruthy();
  await administrador.dispose();
});

test("al modificar algo queda registrado quien lo modifico", async () => {
  const administrador = await comoAdministrador();
  const creada = await administrador.post("/catalogos/minas", {
    data: { nombre: `Mina Cambio ${sello("")}`, area: "Sur", niveles: "" }
  });
  const { id } = await creada.json();

  const antes = await consultar<{ actualizado_por: string | null }>(
    "SELECT actualizado_por_usuario_id::text AS actualizado_por FROM catalogos.mina WHERE id = $1",
    [id]
  );
  expect(antes[0].actualizado_por, "un alta no deberia marcar una actualizacion ajena").toBeNull();

  expect(
    (await administrador.patch(`/catalogos/minas/${id}/estado`, { data: { estado: "INACTIVA" } })).status()
  ).toBe(200);

  const despues = await consultar<{ actualizado_por: string | null }>(
    "SELECT actualizado_por_usuario_id::text AS actualizado_por FROM catalogos.mina WHERE id = $1",
    [id]
  );
  expect(
    despues[0].actualizado_por,
    "el cambio de estado no dejo rastro de quien lo hizo"
  ).toBe(await identificadorDeAdministrador());
  await administrador.dispose();
});

test("un usuario nuevo tambien queda con su autor", async () => {
  const administrador = await comoAdministrador();
  const usuario = sello("qa.autor");
  const creado = await administrador.post("/gobierno/usuarios", {
    data: { usuario, nombre: "QA AUTOR", correo: "", contrasena: contrasenaValida(), id_empleado: "" }
  });
  expect(creado.status()).toBe(201);
  const { id } = await creado.json();

  const filas = await consultar<{ creado_por: string | null }>(
    "SELECT creado_por_usuario_id::text AS creado_por FROM gobierno.usuario WHERE id = $1",
    [id]
  );
  expect(
    filas[0].creado_por,
    "no quedo registrado quien dio de alta al usuario"
  ).toBe(await identificadorDeAdministrador());
  await administrador.dispose();
});
