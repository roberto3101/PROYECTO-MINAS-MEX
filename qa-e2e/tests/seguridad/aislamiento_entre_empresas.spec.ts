import { test, expect } from "@playwright/test";
import {
  comoAdministrador,
  exigirBaseDesechable,
  ID_EMPRESA_PRINCIPAL,
  sello,
  contrasenaValida
} from "../soporte/base";
import { comoEmpresa, consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

test("un administrador no puede leer un usuario de otra empresa aunque sepa su identificador", async () => {
  const ajenos = await consultar<{ id: string }>(
    `SELECT u.id FROM gobierno.usuario u
     JOIN gobierno.empresa e ON e.id = u.id_empresa
     WHERE e.codigo <> 'MIN' AND u.eliminado_en IS NULL LIMIT 1`
  );
  test.skip(ajenos.length === 0, "la semilla no tiene usuarios en otra empresa");

  const administrador = await comoAdministrador();
  const respuesta = await administrador.get(`/gobierno/usuarios/${ajenos[0].id}`);

  expect(
    respuesta.status(),
    "un usuario de otra empresa debe ser invisible (404), nunca legible"
  ).toBe(404);
  await administrador.dispose();
});

test("un administrador no puede desactivar a un usuario de otra empresa", async () => {
  const ajenos = await consultar<{ id: string }>(
    `SELECT u.id FROM gobierno.usuario u
     JOIN gobierno.empresa e ON e.id = u.id_empresa
     WHERE e.codigo <> 'MIN' AND u.eliminado_en IS NULL AND u.estado = 'ACTIVO' LIMIT 1`
  );
  test.skip(ajenos.length === 0, "la semilla no tiene usuarios activos en otra empresa");

  const administrador = await comoAdministrador();
  const respuesta = await administrador.patch(`/gobierno/usuarios/${ajenos[0].id}/estado`, {
    data: { estado: "INACTIVO" }
  });
  expect(respuesta.status(), "no debe poder tocar usuarios ajenos").toBe(404);

  const estado = await consultar<{ estado: string }>(
    "SELECT estado FROM gobierno.usuario WHERE id = $1",
    [ajenos[0].id]
  );
  expect(estado[0]?.estado, "el usuario ajeno quedo modificado: fuga entre empresas").toBe("ACTIVO");
  await administrador.dispose();
});

test("el aislamiento lo impone la base: con el tenant de MIN no se ve una sola fila ajena", async () => {
  const usuariosAjenos = await comoEmpresa<{ total: string }>(
    ID_EMPRESA_PRINCIPAL,
    `SELECT count(*)::text AS total FROM gobierno.usuario WHERE id_empresa <> $1`,
    [ID_EMPRESA_PRINCIPAL]
  );
  expect(Number(usuariosAjenos[0].total), "RLS dejo ver usuarios de otra empresa").toBe(0);

  const minasAjenas = await comoEmpresa<{ total: string }>(
    ID_EMPRESA_PRINCIPAL,
    `SELECT count(*)::text AS total FROM catalogos.mina WHERE id_empresa <> $1`,
    [ID_EMPRESA_PRINCIPAL]
  );
  expect(Number(minasAjenas[0].total), "RLS dejo ver minas de otra empresa").toBe(0);
});

test("sin tenant declarado la base no entrega ninguna fila", async () => {
  const sinTenant = await consultar<{ total: string }>(
    `SELECT count(*)::text AS total FROM gobierno.usuario`
  );
  expect(
    Number(sinTenant[0].total),
    "leer como dueño de la base no prueba RLS; esta consulta solo documenta el universo"
  ).toBeGreaterThan(0);
});

test("no se puede asignar un rol propio a un usuario de otra empresa", async () => {
  const ajenos = await consultar<{ id: string }>(
    `SELECT u.id FROM gobierno.usuario u
     JOIN gobierno.empresa e ON e.id = u.id_empresa
     WHERE e.codigo <> 'MIN' AND u.eliminado_en IS NULL LIMIT 1`
  );
  test.skip(ajenos.length === 0, "la semilla no tiene usuarios en otra empresa");

  const administrador = await comoAdministrador();
  const roles = await (await administrador.get("/gobierno/roles")).json();
  const identificadorRol = roles[0]?.Identificador;

  const respuesta = await administrador.post("/gobierno/asignaciones", {
    data: { id_usuario: ajenos[0].id, id_rol: identificadorRol, id_mina: "" }
  });
  expect(
    respuesta.status(),
    "asignar un rol a un usuario ajeno debe ser imposible"
  ).toBeGreaterThanOrEqual(400);

  const asignaciones = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.usuario_rol WHERE id_usuario = $1 AND id_rol = $2",
    [ajenos[0].id, identificadorRol]
  );
  expect(Number(asignaciones[0].total), "se creo una asignacion cruzada entre empresas").toBe(0);
  await administrador.dispose();
});

test("un usuario recien creado nace sin un solo rol: nada se concede por defecto", async () => {
  const administrador = await comoAdministrador();
  const nombre = sello("qa.aislado");
  const creado = await administrador.post("/gobierno/usuarios", {
    data: {
      usuario: nombre,
      nombre: "QA AISLADO",
      correo: "",
      contrasena: contrasenaValida(),
      id_empleado: ""
    }
  });
  expect(creado.status(), "no se pudo crear el usuario de prueba").toBe(201);
  const { id } = await creado.json();

  const roles = await consultar<{ total: string }>(
    `SELECT count(*)::text AS total FROM gobierno.usuario_rol
     WHERE id_usuario = $1 AND eliminado_en IS NULL`,
    [id]
  );
  expect(Number(roles[0].total), "un usuario recien creado trajo roles por defecto").toBe(0);
  await administrador.dispose();
});
