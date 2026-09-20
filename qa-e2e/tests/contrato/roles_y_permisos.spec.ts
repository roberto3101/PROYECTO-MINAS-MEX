import { test, expect } from "@playwright/test";
import { comoAdministrador, exigirBaseDesechable, sello } from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

test("el catalogo de permisos es global y completo", async () => {
  const administrador = await comoAdministrador();
  const permisos = await (await administrador.get("/gobierno/permisos")).json();

  const enBase = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.permiso WHERE eliminado_en IS NULL AND estado = 'ACTIVO'"
  );
  expect(
    permisos.length,
    "el catalogo que ve la empresa debe coincidir con el catalogo real"
  ).toBe(Number(enBase[0].total));
  await administrador.dispose();
});

test("un rol creado por la empresa nace vacio y no es de sistema", async () => {
  const administrador = await comoAdministrador();
  const codigo = sello("QA_VACIO").toUpperCase();

  const creado = await administrador.post("/gobierno/roles", {
    data: { codigo, descripcion: "Rol creado por la empresa" }
  });
  expect(creado.status()).toBe(201);
  const { id } = await creado.json();

  const filas = await consultar<{ codigo: string; es_sistema: boolean; estado: string }>(
    "SELECT codigo, es_sistema, estado FROM gobierno.rol WHERE id = $1",
    [id]
  );
  expect(filas[0].codigo).toBe(codigo);
  expect(filas[0].es_sistema, "un rol de la empresa no puede nacer como rol de sistema").toBe(false);

  const permisos = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol_permiso WHERE id_rol = $1",
    [id]
  );
  expect(Number(permisos[0].total), "un rol nuevo no debe traer permisos regalados").toBe(0);
  await administrador.dispose();
});

test("conceder un permiso lo deja en la matriz, y concederlo dos veces avisa del conflicto", async () => {
  const administrador = await comoAdministrador();
  const creado = await administrador.post("/gobierno/roles", {
    data: { codigo: sello("QA_CONC").toUpperCase(), descripcion: "Rol para conceder" }
  });
  const { id } = await creado.json();

  const primera = await administrador.post(`/gobierno/roles/${id}/permisos`, {
    data: { permiso: "catalogos.ver" }
  });
  expect(primera.status(), "no se pudo conceder el permiso").toBe(200);

  const enMatriz = await consultar<{ codigo: string }>(
    `SELECT p.codigo FROM gobierno.rol_permiso rp
     JOIN gobierno.permiso p ON p.id = rp.id_permiso
     WHERE rp.id_rol = $1`,
    [id]
  );
  expect(enMatriz.map((f) => f.codigo), "el permiso no quedo en la matriz").toEqual(["catalogos.ver"]);

  const repetida = await administrador.post(`/gobierno/roles/${id}/permisos`, {
    data: { permiso: "catalogos.ver" }
  });
  expect(repetida.status(), "conceder dos veces el mismo permiso debe dar conflicto").toBe(409);

  const cuantos = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol_permiso WHERE id_rol = $1",
    [id]
  );
  expect(Number(cuantos[0].total), "el permiso quedo duplicado en la matriz").toBe(1);
  await administrador.dispose();
});

test("no se puede conceder un permiso que no existe", async () => {
  const administrador = await comoAdministrador();
  const creado = await administrador.post("/gobierno/roles", {
    data: { codigo: sello("QA_FANT").toUpperCase(), descripcion: "Rol para permiso inexistente" }
  });
  const { id } = await creado.json();

  const respuesta = await administrador.post(`/gobierno/roles/${id}/permisos`, {
    data: { permiso: "inventado.total" }
  });
  expect(respuesta.status(), "se acepto un permiso que no esta en el catalogo").toBe(404);

  const cuantos = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol_permiso WHERE id_rol = $1",
    [id]
  );
  expect(Number(cuantos[0].total)).toBe(0);
  await administrador.dispose();
});

test("el rol Administrador es intocable: no se le pueden cambiar los permisos", async () => {
  const administrador = await comoAdministrador();
  const roles = await (await administrador.get("/gobierno/roles")).json();
  const sistema = roles.find((r: any) => r.Codigo === "ADMIN_EMPRESA");
  expect(sistema, "no se encontro el rol de sistema").toBeTruthy();

  const antes = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol_permiso WHERE id_rol = $1",
    [sistema.Identificador]
  );

  const respuesta = await administrador.post(`/gobierno/roles/${sistema.Identificador}/permisos`, {
    data: { permiso: "catalogos.ver" }
  });
  expect(
    respuesta.status(),
    "el rol de sistema debe quedar protegido (403), no aceptar cambios"
  ).toBe(403);

  const despues = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol_permiso WHERE id_rol = $1",
    [sistema.Identificador]
  );
  expect(
    despues[0].total,
    "la matriz del rol de sistema cambio pese al rechazo"
  ).toBe(antes[0].total);
  await administrador.dispose();
});

test("el codigo de rol no se puede repetir dentro de la empresa", async () => {
  const administrador = await comoAdministrador();
  const codigo = sello("QA_REP").toUpperCase();
  const datos = { codigo, descripcion: "Rol repetido" };

  expect((await administrador.post("/gobierno/roles", { data: datos })).status()).toBe(201);
  const repetido = await administrador.post("/gobierno/roles", { data: datos });
  expect(repetido.status(), "se permitio repetir el codigo de rol").toBeGreaterThanOrEqual(400);

  const cuantos = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol WHERE codigo = $1 AND eliminado_en IS NULL",
    [codigo]
  );
  expect(Number(cuantos[0].total), "quedaron dos roles con el mismo codigo").toBe(1);
  await administrador.dispose();
});

test("revocar una asignacion no borra el rol ni su matriz de permisos", async () => {
  const administrador = await comoAdministrador();
  const rol = await administrador.post("/gobierno/roles", {
    data: { codigo: sello("QA_REV").toUpperCase(), descripcion: "Rol para revocar" }
  });
  const { id: identificadorRol } = await rol.json();
  await administrador.post(`/gobierno/roles/${identificadorRol}/permisos`, {
    data: { permiso: "catalogos.ver" }
  });

  const usuarios = await (await administrador.get("/gobierno/usuarios?limite=5")).json();
  const objetivo = usuarios.Elementos[0];

  const asignado = await administrador.post("/gobierno/asignaciones", {
    data: { id_usuario: objetivo.Identificador, id_rol: identificadorRol, id_mina: "" }
  });
  expect(asignado.status()).toBe(201);
  const { id: identificadorAsignacion } = await asignado.json();

  expect((await administrador.delete(`/gobierno/asignaciones/${identificadorAsignacion}`)).status()).toBe(200);

  const repetida = await administrador.delete(`/gobierno/asignaciones/${identificadorAsignacion}`);
  expect(repetida.status(), "revocar dos veces debe avisar del conflicto").toBe(409);

  const rolVive = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol WHERE id = $1 AND eliminado_en IS NULL",
    [identificadorRol]
  );
  expect(Number(rolVive[0].total), "revocar una asignacion borro el rol").toBe(1);

  const matrizVive = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.rol_permiso WHERE id_rol = $1",
    [identificadorRol]
  );
  expect(Number(matrizVive[0].total), "revocar una asignacion borro los permisos del rol").toBe(1);
  await administrador.dispose();
});
